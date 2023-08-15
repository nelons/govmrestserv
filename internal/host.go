package internal

import (
	"context"
	"net/http"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/nelons/vsphere-rest-server/pkg/swagger/server/restapi/operations"
	"github.com/vmware/govmomi/vim25/mo"
)

func write_hosts(items []mo.HostSystem, props []string, rw http.ResponseWriter) {
	rw.WriteHeader(200)

	jsonObj := gabs.New()
	jsonObj.Set(int64(len(items)), "count")
	jsonObj.Array("results")

	for _, item := range items {
		objData, err := serialise_object(item, nil, props) //serialise_object_json(vm, nil)

		if err == nil {
			// Add to the results array
			jsonObj.ArrayAppend(objData, "results")
		}
	}

	out := jsonObj.String()
	rw.Write([]byte(out))
}

func get_vsphere_get_host(user operations.VSphereGetAllHostsSummaryParams) middleware.Responder {
	vc, err := validate_request(user.HTTPRequest, user.VRSToken, user.Vcenter)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetAllHostsSummaryBadRequest().WithPayload(&body)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var hosts []mo.HostSystem
	err = vcenter_get_objects(vc.client, ctx, ObjectType_HostSystem, []string{}, &hosts)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetAllHostsSummaryBadRequest().WithPayload(&body)
	}

	return middleware.ResponderFunc(func(rw http.ResponseWriter, p runtime.Producer) {
		write_hosts(hosts, nil, rw)
	})
}

func get_vsphere_get_host_byname(user operations.VSphereGetHostByNameParams) middleware.Responder {
	vc, err := validate_request(user.HTTPRequest, user.VRSToken, user.Vcenter)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetHostByNameBadRequest().WithPayload(&body)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var props []string
	if user.Props != nil {
		props = strings.Split(*user.Props, ",")
	}

	var hosts []mo.HostSystem
	err = vcenter_get_object_byname(vc.client, ctx, ObjectType_HostSystem, user.Hostname, props, &hosts)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetHostByNameBadRequest().WithPayload(&body)
	}

	return middleware.ResponderFunc(func(rw http.ResponseWriter, p runtime.Producer) {
		write_hosts(hosts, props, rw)
	})
}

func get_vsphere_get_host_byref(user operations.VSphereGetHostByMoRefParams) middleware.Responder {
	vc, err := validate_request(user.HTTPRequest, user.VRSToken, user.Vcenter)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetHostByMoRefBadRequest().WithPayload(&body)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var props []string
	if user.Props != nil {
		props = strings.Split(*user.Props, ",")
	}

	var hosts []mo.HostSystem
	err = vcenter_get_object_byref(vc.client, ctx, ObjectType_HostSystem, user.Moref, props, &hosts)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetHostByMoRefBadRequest().WithPayload(&body)
	}

	return middleware.ResponderFunc(func(rw http.ResponseWriter, p runtime.Producer) {
		write_hosts(hosts, props, rw)
	})
}
