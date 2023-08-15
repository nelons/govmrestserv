package internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/Jeffail/gabs/v2"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/nelons/vsphere-rest-server/pkg/swagger/server/restapi/operations"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type vcenter_connection struct {
	URL      string
	Username string
	Name     string

	access_mutex sync.Mutex
	client       *vim25.Client
}

var connections []*vcenter_connection
var connections_mutex sync.Mutex

type ObjectType string

const (
	ObjectType_VirtualMachine         ObjectType = "VirtualMachine"
	ObjectType_HostSystem             ObjectType = "HostSystem"
	ObjectType_Datastore              ObjectType = "Datastore"
	ObjectType_Datacenter             ObjectType = "Datacenter"
	ObjectType_ClusterComputeResource ObjectType = "ClusterComputeResource"
	ObjectType_ResourcePool           ObjectType = "ResourcePool"
	ObjectType_StoragePod             ObjectType = "StoragePod"
	ObjectType_Network                ObjectType = "Network"
)

func (ot ObjectType) String() string {
	return string(ot)
}

func vcenter_login(ctx context.Context, vc_url, user, pwd string, allow_insecure bool) (*vim25.Client, error) {
	u, err := url.Parse(vc_url)
	if err != nil {
		return nil, err
	}

	if u == nil {
		return nil, errors.New("failed to parse url '" + vc_url + "'")
	}

	u.User = url.UserPassword(user, pwd)

	s := &cache.Session{
		URL:      u,
		Insecure: allow_insecure,
	}

	client := new(vim25.Client)
	err = s.Login(ctx, client, nil)
	if err != nil {
		fmt.Printf("Login Failure - error %v", err)
		return nil, err
	}

	return client, nil
}

func vcenter_get_objects(client *vim25.Client, ctx context.Context, object_type ObjectType, props []string, objects interface{}) error {
	m := view.NewManager(client)

	v, err := m.CreateContainerView(ctx, client.ServiceContent.RootFolder, []string{object_type.String()}, true)
	if err != nil {
		return err
	}
	defer v.Destroy(ctx)

	err = v.Retrieve(ctx, []string{object_type.String()}, props, objects)
	if err != nil {
		return err
	}

	return nil
}
func vcenter_get_object_byname(client *vim25.Client, ctx context.Context, object_type ObjectType, object_name string, props []string, objects interface{}) error {
	m := view.NewManager(client)

	v, err := m.CreateContainerView(ctx, client.ServiceContent.RootFolder, []string{object_type.String()}, true)
	if err != nil {
		return err
	}

	defer v.Destroy(ctx)

	err = v.RetrieveWithFilter(ctx, []string{object_type.String()}, props, objects, property.Filter{"name": object_name})
	if err != nil {
		return err
	}

	return nil
}

func vcenter_get_object_byref(client *vim25.Client, ctx context.Context, object_type ObjectType, object_ref string, props []string, objects interface{}) error {
	m := view.NewManager(client)

	v, err := m.CreateContainerView(ctx, client.ServiceContent.RootFolder, []string{object_type.String()}, true)
	if err != nil {
		return err
	}

	defer v.Destroy(ctx)

	var moref types.ManagedObjectReference
	moref.Type = object_type.String()
	moref.Value = object_ref

	fmt.Printf("Attemping to get moref %v\n", moref)

	err = v.RetrieveWithFilter(ctx, []string{object_type.String()}, props, objects, property.Filter{"Self": moref})
	if err != nil {
		return err
	}

	return nil
}

/*
Some writing functions
*/
func write_datastores(items []mo.Datastore, rw http.ResponseWriter) {
	rw.WriteHeader(200)

	jsonObj := gabs.New()
	jsonObj.Set(int64(len(items)), "count")
	jsonObj.Array("results")

	for _, item := range items {
		objData, err := serialise_object(item, nil, []string{}) //serialise_object_json(vm, nil)

		if err == nil {
			// Add to the results array
			jsonObj.ArrayAppend(objData, "results")
		}
	}

	out := jsonObj.String()
	rw.Write([]byte(out))
}

/*
These are httpserv functions that are more suited to vsphere
Might move them later idk
*/
/*func test_write_slice(objects interface{}, rw http.ResponseWriter) {
	rw.WriteHeader(200)

	jsonObj := gabs.New()
	jsonObj.Set(int64(len(items)), "count")
	jsonObj.Array("results")

	for _, item := range items {
		objData, err := serialise_object(item, nil) //serialise_object_json(vm, nil)

		if err == nil {
			// Add to the results array
			jsonObj.ArrayAppend(objData, "results")
		}
	}

	out := jsonObj.String()
	rw.Write([]byte(out))
}*/

func get_vsphere_get_datastore(user operations.VSphereGetAllDatastoresParams) middleware.Responder {
	vc, err := validate_request(user.HTTPRequest, user.VRSToken, user.Vcenter)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetAllDatastoresBadRequest().WithPayload(&body)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var dss []mo.Datastore
	err = vcenter_get_objects(vc.client, ctx, ObjectType_Datastore, []string{}, &dss)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetAllDatastoresBadRequest().WithPayload(&body)
	}

	return middleware.ResponderFunc(func(rw http.ResponseWriter, p runtime.Producer) {
		write_datastores(dss, rw)
	})
}

func write_networks(items []mo.Network, rw http.ResponseWriter) {
	rw.WriteHeader(200)

	jsonObj := gabs.New()
	jsonObj.Set(int64(len(items)), "count")
	jsonObj.Array("results")

	for _, item := range items {
		objData, err := serialise_object(item, nil, []string{}) //serialise_object_json(vm, nil)

		if err == nil {
			// Add to the results array
			jsonObj.ArrayAppend(objData, "results")
		}
	}

	out := jsonObj.String()
	rw.Write([]byte(out))
}

func get_vsphere_get_network(user operations.VSphereGetAllNetworksParams) middleware.Responder {
	vc, err := validate_request(user.HTTPRequest, user.VRSToken, user.Vcenter)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetAllNetworksBadRequest().WithPayload(&body)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var networks []mo.Network
	err = vcenter_get_objects(vc.client, ctx, ObjectType_Network, []string{}, &networks)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetAllNetworksBadRequest().WithPayload(&body)
	}

	return middleware.ResponderFunc(func(rw http.ResponseWriter, p runtime.Producer) {
		write_networks(networks, rw)
	})
}

func write_datacenters(items []mo.Datacenter, rw http.ResponseWriter) {
	rw.WriteHeader(200)

	jsonObj := gabs.New()
	jsonObj.Set(int64(len(items)), "count")
	jsonObj.Array("results")

	for _, item := range items {
		objData, err := serialise_object(item, nil, []string{}) //serialise_object_json(vm, nil)

		if err == nil {
			// Add to the results array
			jsonObj.ArrayAppend(objData, "results")
		}
	}

	out := jsonObj.String()
	rw.Write([]byte(out))
}

func get_vsphere_get_datacenter(user operations.VSphereGetAllDatacentersParams) middleware.Responder {
	vc, err := validate_request(user.HTTPRequest, user.VRSToken, user.Vcenter)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetAllDatacentersBadRequest().WithPayload(&body)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var datacenters []mo.Datacenter
	err = vcenter_get_objects(vc.client, ctx, ObjectType_Datacenter, []string{}, &datacenters)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetAllDatacentersBadRequest().WithPayload(&body)
	}

	return middleware.ResponderFunc(func(rw http.ResponseWriter, p runtime.Producer) {
		write_datacenters(datacenters, rw)
	})
}

func write_clusters(items []mo.ClusterComputeResource, rw http.ResponseWriter) {
	rw.WriteHeader(200)

	jsonObj := gabs.New()
	jsonObj.Set(int64(len(items)), "count")
	jsonObj.Array("results")

	for _, item := range items {
		objData, err := serialise_object(item, nil, []string{}) //serialise_object_json(vm, nil)

		if err == nil {
			// Add to the results array
			jsonObj.ArrayAppend(objData, "results")
		}
	}

	out := jsonObj.String()
	rw.Write([]byte(out))
}

func get_vsphere_get_cluster(user operations.VSphereGetAllClustersParams) middleware.Responder {
	vc, err := validate_request(user.HTTPRequest, user.VRSToken, user.Vcenter)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetAllClustersBadRequest().WithPayload(&body)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var clusters []mo.ClusterComputeResource
	err = vcenter_get_objects(vc.client, ctx, ObjectType_ClusterComputeResource, []string{}, &clusters)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetAllClustersBadRequest().WithPayload(&body)
	}

	return middleware.ResponderFunc(func(rw http.ResponseWriter, p runtime.Producer) {
		write_clusters(clusters, rw)
	})
}

func write_storagepods(items []mo.StoragePod, rw http.ResponseWriter) {
	rw.WriteHeader(200)

	jsonObj := gabs.New()
	jsonObj.Set(int64(len(items)), "count")
	jsonObj.Array("results")

	for _, item := range items {
		objData, err := serialise_object(item, nil, []string{}) //serialise_object_json(vm, nil)

		if err == nil {
			// Add to the results array
			jsonObj.ArrayAppend(objData, "results")
		}
	}

	out := jsonObj.String()
	rw.Write([]byte(out))
}

func get_vsphere_get_storagepod(user operations.VSphereGetAllStoragePodsParams) middleware.Responder {
	vc, err := validate_request(user.HTTPRequest, user.VRSToken, user.Vcenter)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetAllStoragePodsBadRequest().WithPayload(&body)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var pods []mo.StoragePod
	err = vcenter_get_objects(vc.client, ctx, ObjectType_StoragePod, []string{}, &pods)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetAllStoragePodsBadRequest().WithPayload(&body)
	}

	return middleware.ResponderFunc(func(rw http.ResponseWriter, p runtime.Producer) {
		write_storagepods(pods, rw)
	})
}

func write_resourcepools(items []mo.ResourcePool, rw http.ResponseWriter) {
	rw.WriteHeader(200)

	jsonObj := gabs.New()
	jsonObj.Set(int64(len(items)), "count")
	jsonObj.Array("results")

	for _, item := range items {
		objData, err := serialise_object(item, nil, []string{}) //serialise_object_json(vm, nil)

		if err == nil {
			// Add to the results array
			jsonObj.ArrayAppend(objData, "results")
		}
	}

	out := jsonObj.String()
	rw.Write([]byte(out))
}

func get_vsphere_get_resourcepool(user operations.VSphereGetAllResourcePoolParams) middleware.Responder {
	vc, err := validate_request(user.HTTPRequest, user.VRSToken, user.Vcenter)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetAllResourcePoolBadRequest().WithPayload(&body)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var pools []mo.ResourcePool
	err = vcenter_get_objects(vc.client, ctx, ObjectType_ResourcePool, []string{}, &pools)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetAllResourcePoolBadRequest().WithPayload(&body)
	}

	return middleware.ResponderFunc(func(rw http.ResponseWriter, p runtime.Producer) {
		write_resourcepools(pools, rw)
	})
}
