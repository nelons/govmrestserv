package internal

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	"github.com/nelons/vsphere-rest-server/pkg/swagger/server/models"
	"github.com/nelons/vsphere-rest-server/pkg/swagger/server/restapi"
	"github.com/nelons/vsphere-rest-server/pkg/swagger/server/restapi/operations"
)

type json_wrapper struct {
	TypeName string
	Object   any
}

func InitialiseServer() {
	cert_exists := file_exists("cert.cer")
	key_exists := file_exists("cert.key")

	if !cert_exists || !key_exists {
		if cert_exists {
			err := os.Remove("cert.cer")
			if err != nil {
				log.Fatal("Could not remove certificate file.")
			}
			cert_exists = false
		}

		if key_exists {
			err := os.Remove("cert.key")
			if err != nil {
				log.Fatal("Could not remove private key.")
			}
			key_exists = false
		}
	}

	if !cert_exists && !key_exists {
		generate_selfsigned_certificate()
	}
}

func StartServer() {
	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewVSphereAPI(swaggerSpec)
	server := restapi.NewServer(api)
	defer func() {
		if err := server.Shutdown(); err != nil {
			log.Fatalln(err)
		}
		log.Println("Server shutdown !")
	}()

	server.TLSPort = 8443
	server.TLSCertificate = "cert.cer"
	server.TLSCertificateKey = "cert.key"
	server.TLSCACertificate = ""

	api.SessionRegisterHandler = operations.SessionRegisterHandlerFunc(post_session_register)
	api.SessionListHandler = operations.SessionListHandlerFunc(get_session_list)
	api.VSphereConnectHandler = operations.VSphereConnectHandlerFunc(post_vsphere_connect)
	api.VSphereListConnectionsHandler = operations.VSphereListConnectionsHandlerFunc(get_vSphere_list)
	api.VSphereGetAllVMSummaryHandler = operations.VSphereGetAllVMSummaryHandlerFunc(get_vsphere_get_vms)
	api.VSphereGetVMByNameHandler = operations.VSphereGetVMByNameHandlerFunc(get_vsphere_get_vm_byname)
	api.VSphereGetVMByMoRefHandler = operations.VSphereGetVMByMoRefHandlerFunc(get_vsphere_get_vm_bymoref)
	api.VSphereGetAllHostsSummaryHandler = operations.VSphereGetAllHostsSummaryHandlerFunc(get_vsphere_get_host)

	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}
}

func TestServer() {
	// Testing, connect to a vcenter.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := vcenter_login(ctx, "https://127.0.0.1:8005/sdk", "user", "pass", true)
	if err != nil {
		return
	}

	vms, err := vcenter_getvm_byname(client, ctx, "mx-wp-db16-dr")
	if err == nil {
		for _, vm := range vms {
			fmt.Printf("Found VM by Name: '%v'\n", vm.Name)
		}
	}

	items, err := vcenter_getvm_bymoref(client, ctx, "vm-8567")
	if err == nil {
		for _, vm := range items {
			fmt.Printf("Found VM by Ref: '%v'\n", vm.Name)
		}
	} else {
		fmt.Printf("Error: %v\n", err.Error())
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

/*func write_raw_collection(items []any) []interface{} {
	out := make([]interface{}, len(items))
	i := 0
	for _, vm := range items {
		out[i] = &vm
		i++
	}

	return out
}*/

func create_badrequesterror(msg string) models.BadRequestError {
	var body models.BadRequestError
	body.Error = msg
	return body
}
