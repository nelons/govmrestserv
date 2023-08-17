package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/Jeffail/gabs/v2"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	"github.com/jessevdk/go-flags"
	"github.com/nelons/vsphere-rest-server/pkg/swagger/server/models"
	"github.com/nelons/vsphere-rest-server/pkg/swagger/server/restapi"
	"github.com/nelons/vsphere-rest-server/pkg/swagger/server/restapi/operations"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
)

type Configuration struct {
	Port             int    `json:"port"`
	Certificate_file string `json:"certificate_file"`
	Certificate_key  string `json:"certificate_key"`
}

var server *restapi.Server

func InitialiseServer() error {
	var config Configuration

	// Load the configuration
	if file_exists("config.json") {
		configFile, err := os.Open("config.json")
		if err != nil {
			fmt.Println(err.Error())

		} else {
			defer configFile.Close()
			jsonParser := json.NewDecoder(configFile)
			jsonParser.Decode(&config)

		}
	}

	if config.Port == 0 {
		config.Port = 8443
	}

	cert_exists := file_exists(config.Certificate_file)
	key_exists := file_exists(config.Certificate_key)

	if !cert_exists || !key_exists {
		cert_exists := file_exists("cert.cer")
		key_exists := file_exists("cert.key")

		if cert_exists {
			err := os.Remove("cert.cer")
			if err != nil {
				log.Fatal("Could not remove certificate file.")
				return err

			}
			cert_exists = false
		}

		if key_exists {
			err := os.Remove("cert.key")
			if err != nil {
				log.Fatal("Could not remove private key.")
				return err

			}
			key_exists = false
		}

		if !cert_exists && !key_exists {
			generate_selfsigned_certificate()
			config.Certificate_file = "cert.cer"
			config.Certificate_key = "cert.key"

			// TODO: save out JSON

		} else {
			return errors.New("Failed to generate self-signed certificate.")

		}
	}

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		log.Fatalln(err)
		return err
	}

	api := operations.NewVSphereAPI(swaggerSpec)

	api.SessionRegisterHandler = operations.SessionRegisterHandlerFunc(post_session_register)
	api.SessionListHandler = operations.SessionListHandlerFunc(get_session_list)
	api.VSphereConnectHandler = operations.VSphereConnectHandlerFunc(post_vsphere_connect)
	api.VSphereListConnectionsHandler = operations.VSphereListConnectionsHandlerFunc(get_vSphere_list)

	// VMs
	api.VSphereGetAllVMSummaryHandler = operations.VSphereGetAllVMSummaryHandlerFunc(get_vsphere_get_vms)
	api.VSphereGetVMByNameHandler = operations.VSphereGetVMByNameHandlerFunc(get_vsphere_get_vm_byname)
	api.VSphereGetVMByMoRefHandler = operations.VSphereGetVMByMoRefHandlerFunc(get_vsphere_get_vm_bymoref)
	api.VSphereChangeVMPowerStateHandler = operations.VSphereChangeVMPowerStateHandlerFunc(post_vsphere_vm_power)

	// Hosts
	api.VSphereGetAllHostsSummaryHandler = operations.VSphereGetAllHostsSummaryHandlerFunc(get_vsphere_get_host)
	api.VSphereGetHostByNameHandler = operations.VSphereGetHostByNameHandlerFunc(get_vsphere_get_host_byname)
	api.VSphereGetHostByMoRefHandler = operations.VSphereGetHostByMoRefHandlerFunc(get_vsphere_get_host_byref)

	// Datastores
	api.VSphereGetAllDatastoresHandler = operations.VSphereGetAllDatastoresHandlerFunc(get_vsphere_get_datastore)

	// Networks
	api.VSphereGetAllNetworksHandler = operations.VSphereGetAllNetworksHandlerFunc(get_vsphere_get_network)

	// vCenter Stuff
	api.VSphereGetAllDatacentersHandler = operations.VSphereGetAllDatacentersHandlerFunc(get_vsphere_get_datacenter)
	api.VSphereGetAllClustersHandler = operations.VSphereGetAllClustersHandlerFunc(get_vsphere_get_cluster)
	api.VSphereGetAllResourcePoolHandler = operations.VSphereGetAllResourcePoolHandlerFunc(get_vsphere_get_resourcepool)
	api.VSphereGetAllStoragePodsHandler = operations.VSphereGetAllStoragePodsHandlerFunc(get_vsphere_get_storagepod)

	server = restapi.NewServer(api)

	server.TLSPort = config.Port
	server.TLSCertificate = flags.Filename(config.Certificate_file)
	server.TLSCertificateKey = flags.Filename(config.Certificate_key)
	server.TLSCACertificate = ""
	return nil
}

var ticker *time.Ticker
var quit chan struct{}

func StartServer() {
	ticker = time.NewTicker(time.Duration(60) * time.Second)
	quit = make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				// do stuff
				checkClientConnections()

			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	log.Println("Server starting.")
	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}
}

func ShutdownServer() {
	err := server.Shutdown()
	if err != nil {
		log.Fatalln(err)
	}

	close(quit)
}

func Test_PrintObject(obj any) {
	prettyJSON, err := json.MarshalIndent(obj, "", "  ")
	if err == nil {
		fmt.Printf("%v\n", string(prettyJSON))
	}
}

/*
This tests the serialisation output function
and should not output unrequested values.
*/
func Test_OutputObject(obj any, props []string) {
	fmt.Printf("Outputing properties: %v\n", props)

	jsonObj := gabs.New()

	objData, err := serialise_object(obj, nil, props)

	if err == nil {
		jsonObj.Set(objData)
	}

	out := jsonObj.String()
	fmt.Printf("%v\n", out)
}

func TestServer(vcenter_sdk, username, password string) {
	// Testing, connect to a vcenter.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := vcenter_login(ctx, vcenter_sdk, username, password, true)
	if err != nil {
		return
	}

	fmt.Printf("Connected to %v.\n", vcenter_sdk)
	about := client.ServiceContent.About
	Test_PrintObject(about)

	// fork code path between ESXi Host and vCenter
	// as some entities below are only applicable to vCenter.
	if about.ApiType == "VirtualCenter" {
		fmt.Println("VirtualCenter detected.")

		// Try and get datacenters
		fmt.Println("Datacenters:")
		var dcs []mo.Datacenter
		err = vcenter_get_objects(client, ctx, "Datacenter", []string{}, &dcs)
		if err == nil {
			for _, dc := range dcs {
				fmt.Printf(": %v\n", dc.Name)
			}
		}

		// Compute Clusters
		fmt.Println("ClusterComputeResource:")
		var ccr []mo.ClusterComputeResource
		err = vcenter_get_objects(client, ctx, "ClusterComputeResource", []string{}, &ccr)
		if err == nil {
			for _, cluster := range ccr {
				fmt.Printf(": %v\n", cluster.Name)

				// TODO: get hosts for this cluster
			}
		}

		// Resource Pools
		fmt.Println("Resource Pools:")
		var rps []mo.ResourcePool
		err = vcenter_get_objects(client, ctx, "ResourcePool", []string{}, &rps)
		if err == nil {
			for _, rp := range rps {
				fmt.Printf(": %v\n", rp.Name)
			}
		}

		// StoragePods
		fmt.Println("StoragePods:")
		var pods []mo.StoragePod
		err = vcenter_get_objects(client, ctx, "StoragePod", []string{}, &pods)
		if err == nil {
			for _, pod := range pods {
				fmt.Printf(": %v\n", pod.Name)

				fmt.Printf("-- %v\n", pod.ChildType)

				for _, ds := range pod.ChildEntity {
					fmt.Printf("-- %v\n", ds.String())
				}
			}
		}

	} else {
		fmt.Println("ESX Host detected.")

		// ComputeResource ?
		fmt.Println("ComputeResource:")
		var crs []mo.ClusterComputeResource
		err = vcenter_get_objects(client, ctx, "ComputeResource", []string{}, &crs)
		if err == nil {
			for _, cluster := range crs {
				fmt.Printf(": %v\n", cluster.Name)
			}
		}
	}

	// Datastores
	fmt.Println("Datastores:")
	var dss []mo.Datastore
	err = vcenter_get_objects(client, ctx, "Datastore", []string{}, &dss)
	if err == nil {
		for _, ds := range dss {
			fmt.Printf(": %v\n", ds.Name)
		}
	}

	// Networks
	fmt.Println("Networks:")
	var networks []mo.Network
	err = vcenter_get_objects(client, ctx, "Network", []string{}, &networks)
	if err == nil {
		for _, network := range networks {
			fmt.Printf(": %v\n", network.Name)
		}
	}

	// Get all the hosts
	var hosts []mo.HostSystem
	err = vcenter_get_objects(client, ctx, "HostSystem", []string{}, &hosts)
	if err == nil {
		fmt.Printf("Found %v hosts.\n", len(hosts))

		for _, host := range hosts {
			fmt.Printf("Host: %v\n", host.Name)
		}
	}

	// Get the list of VMs.
	//all_vms, err := vcenter_getvms_summary(client, ctx)
	var all_vms []mo.VirtualMachine
	err = vcenter_get_objects(client, ctx, ObjectType_VirtualMachine, []string{"summary"}, &all_vms)
	if err == nil {
		// Pick a VM at random to get by name
		max_vms := len(all_vms)
		if max_vms > 0 {
			fmt.Printf("Retrieved %v VMs.\n", max_vms)

			vm_position := rand.Intn(max_vms - 1)
			vm := all_vms[vm_position]

			fmt.Printf("Looking for a VM with the name '%v'\n", vm.Summary.Config.Name)

			var vms []mo.VirtualMachine
			err = vcenter_get_object_byname(client, ctx, ObjectType_VirtualMachine, vm.Summary.Config.Name, []string{"Self"}, &vms)
			if err == nil {
				for _, item := range vms {
					fmt.Printf("Found VM by Name: '%v', the reference is '%v'\n", vm.Summary.Config.Name, item.Self.Value)
				}

			} else {
				fmt.Printf("Error getting VM by name: %v\n", err.Error())

			}

			// Pick a VM at random to get by reference
			vm_position = rand.Intn(max_vms - 1)
			vm = all_vms[vm_position]

			fmt.Printf("Looking for a VM with the reference '%v'\n", vm.Self.Value)
			vms = nil
			err = vcenter_get_object_byref(client, ctx, ObjectType_VirtualMachine, vm.Self.Value, []string{"summary"}, &vms)
			if err == nil {
				for _, item := range vms {
					fmt.Printf("Found VM by Ref: '%v', the name is '%v'\n", vm.Self.Value, item.Summary.Config.Name)
				}

			} else {
				fmt.Printf("Error getting VM by reference: %v\n", err.Error())

			}

			mobj := object.NewVirtualMachine(client, vm.Self)
			power_state, err := mobj.PowerState(ctx)
			if err == nil {
				fmt.Printf("Current power state: %v\n", power_state)

				if power_state == "poweredOff" {
					mobj.PowerOn(ctx)

				} else if power_state == "poweredOn" {
					mobj.PowerOff(ctx)

				}
			}

			// Get some VM properties.
			vm_position = rand.Intn(max_vms - 1)
			vm = all_vms[vm_position]
			vms = nil

			fmt.Println("Testing serialisation of a VM with explicitly requested properties.")
			props := []string{"Self", "alarmActionsEnabled", "name"}
			err = vcenter_get_object_byref(client, ctx, ObjectType_VirtualMachine, vm.Self.Value, props, &vms)
			if err == nil {
				if len(vms) == 1 {
					Test_OutputObject(vms[0], props)
				}
			}
		}

	} else {
		fmt.Println("Failed to get Virtual Machines: " + err.Error())

	}
}

func post_session_register(user operations.SessionRegisterParams) middleware.Responder {
	secret := *user.RequestBody.Secret

	client, err := registerNewSession(user.HTTPRequest.RemoteAddr, secret)
	if err != nil {
		var error_body operations.SessionRegisterInternalServerErrorBody
		error_body.Error = err.Error()
		return operations.NewSessionRegisterInternalServerError().WithPayload(&error_body)
	}

	log.Printf("Session registered from %v with token %v.\n", user.HTTPRequest.RemoteAddr, client.Token)

	var ok_body operations.SessionRegisterOKBody
	ok_body.Token = client.Token
	return operations.NewSessionRegisterOK().WithPayload(&ok_body)
}

func get_session_list(user operations.SessionListParams) middleware.Responder {
	clients_mutex.Lock()
	out := make([]*operations.SessionListOKBodyItems0, len(clients))
	for i := 0; i < len(clients); i++ {
		var cl operations.SessionListOKBodyItems0
		cl.Host = clients[i].Host
		cl.Secret = clients[i].Secret

		clients[i].access_mutex.Lock()
		cl.LastAccess = clients[i].LastAccess.Format(time.DateTime)

		cl.Connections = make([]*operations.SessionListOKBodyItems0ConnectionsItems0, len(clients[i].Connections))
		for k := 0; k < cap(clients[i].Connections); k++ {
			var conn operations.SessionListOKBodyItems0ConnectionsItems0
			entry := clients[i].Connections[k]
			conn.URL = entry.URL
			conn.Username = entry.Username
			cl.Connections[i] = &conn
		}

		clients[i].access_mutex.Unlock()

		out[i] = &cl
	}
	clients_mutex.Unlock()

	return operations.NewSessionListOK().WithPayload(out)
}

func post_vsphere_connect(user operations.VSphereConnectParams) middleware.Responder {
	// assert the token is valid
	client, err := verifyClientAcess(user.HTTPRequest.RemoteAddr, user.VRSToken)
	if err != nil {
		return operations.NewVSphereConnectUnauthorized()
	}

	// check the URL is in a proper format
	_, err = url.ParseRequestURI(*user.RequestBody.URL)
	if err != nil {
		// Return error bad argument
		var body operations.VSphereConnectBadRequestBody
		body.Error = err.Error()
		return operations.NewVSphereConnectBadRequest().WithPayload(&body)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	connection := registerClientvCenterConnection(client, *user.RequestBody.URL, *user.RequestBody.Username)
	connection.access_mutex.Lock()
	if connection.client == nil {
		vc, err := vcenter_login(ctx, *user.RequestBody.URL, *user.RequestBody.Username, *user.RequestBody.Password, user.RequestBody.AllowInsecure)
		if err != nil {
			var body operations.VSphereConnectBadRequestBody
			body.Error = err.Error()
			connection.access_mutex.Unlock()
			return operations.NewVSphereConnectBadRequest().WithPayload(&body)
		}

		connection.client = vc
	}
	connection.access_mutex.Unlock()

	// Return some basic information about the target
	about := connection.client.ServiceContent.About
	var body operations.VSphereConnectOKBody
	body.Name = connection.Name
	body.About = about
	return operations.NewVSphereConnectOK().WithPayload(&body)
}

func get_vSphere_list(user operations.VSphereListConnectionsParams) middleware.Responder {
	_, err := verifyClientAcess(user.HTTPRequest.RemoteAddr, user.VRSToken)
	if err != nil {
		return operations.NewVSphereListConnectionsUnauthorized()
	}

	connections_mutex.Lock()
	out := make([]*operations.VSphereListConnectionsOKBodyItems0, cap(connections))

	for i := 0; i < cap(connections); i++ {
		var conn operations.VSphereListConnectionsOKBodyItems0
		conn.URL = connections[i].URL
		conn.Username = connections[i].Username
		conn.Name = connections[i].Name
		out[i] = &conn
	}
	connections_mutex.Unlock()
	return operations.NewVSphereListConnectionsOK().WithPayload(out)
}

func validate_request(request *http.Request, token string, target string) (*vcenter_connection, error) {
	if request == nil {
		return nil, errors.New("Invalid http request.")
	}

	if len(token) == 0 {
		return nil, errors.New("Need to supply a token in the headers.")
	}

	if len(target) == 0 {
		return nil, errors.New("Need to supply a target vcenter/esxi in the path.")
	}

	client, err := verifyClientAcess(request.RemoteAddr, token)
	if err != nil {
		return nil, err
	}

	vc := getClientvCenterConnectionByName(client, target)
	if vc == nil {
		return nil, errors.New("The target is unknown.")
	}

	return vc, nil
}

func create_badrequesterror(msg string) models.BadRequestError {
	var body models.BadRequestError
	body.Error = msg
	return body
}
