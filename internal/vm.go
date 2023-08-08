package internal

import (
	"context"
	"net/http"

	"github.com/Jeffail/gabs/v2"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/nelons/vsphere-rest-server/pkg/swagger/server/models"
	"github.com/nelons/vsphere-rest-server/pkg/swagger/server/restapi/operations"
	"github.com/vmware/govmomi/vim25/mo"
)

/*
Output Functions
*/
func write_virtualmachines(items []mo.VirtualMachine, rw http.ResponseWriter) {
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
}

func get_raw_virtualmachines(items []mo.VirtualMachine) []interface{} {
	out := make([]interface{}, len(items))
	i := 0
	for _, vm := range items {
		out[i] = &vm
		i++
	}

	return out
}

/*
	This contains functions related to the /vm of the server.
*/

func get_vsphere_get_vms(user operations.VSphereGetAllVMSummaryParams) middleware.Responder {
	vc, err := validate_request(user.HTTPRequest, user.VRSToken, user.Vcenter)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetAllVMSummaryBadRequest().WithPayload(&body)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vms, err := vcenter_getvms_summary(vc.client, ctx)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetAllVMSummaryBadRequest().WithPayload(&body)
	}

	out := make([]*operations.VSphereGetAllVMSummaryOKBodyResultsItems0, len(vms))
	for i := 0; i < len(vms); i++ {
		vm := vms[i]
		var item operations.VSphereGetAllVMSummaryOKBodyResultsItems0

		item.Name = vm.Summary.Config.Name
		item.CPU = int64(vm.Summary.Config.NumCpu)
		item.MemoryMB = int64(vm.Summary.Config.MemorySizeMB)
		item.Powerstate = string(vm.Summary.Runtime.PowerState)
		item.Moref = vm.Self.String()
		item.HardwareVersion = vm.Summary.Config.HwVersion
		item.NumberDisks = int64(vm.Summary.Config.NumVirtualDisks)
		item.NumberNICs = int64(vm.Summary.Config.NumEthernetCards)
		item.GuestFullName = vm.Summary.Config.GuestFullName
		item.Status = string(vm.Summary.OverallStatus)

		out[i] = &item
	}

	var body operations.VSphereGetAllVMSummaryOKBody
	body.Count = int64(len(vms))
	body.Results = out

	return operations.NewVSphereGetAllVMSummaryOK().WithPayload(&body)
}

func get_vsphere_get_vm_byname(user operations.VSphereGetVMByNameParams) middleware.Responder {
	vc, err := validate_request(user.HTTPRequest, user.VRSToken, user.Vcenter)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetVMByNameBadRequest().WithPayload(&body)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vms, err := vcenter_getvm_byname(vc.client, ctx, user.Vmname)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetVMByNameBadRequest().WithPayload(&body)
	}

	raw := user.Raw != nil && *user.Raw
	if raw {
		var body models.ObjectCollection
		body.Count = int64(len(vms))
		body.Results = get_raw_virtualmachines(vms)
		return operations.NewVSphereGetVMByNameOK().WithPayload(&body)
	}

	return middleware.ResponderFunc(func(rw http.ResponseWriter, p runtime.Producer) {
		write_virtualmachines(vms, rw)
	})
}

func get_vsphere_get_vm_bymoref(user operations.VSphereGetVMByMoRefParams) middleware.Responder {
	vc, err := validate_request(user.HTTPRequest, user.VRSToken, user.Vcenter)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetVMByMoRefBadRequest().WithPayload(&body)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vms, err := vcenter_getvm_bymoref(vc.client, ctx, user.Moref)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetVMByMoRefBadRequest().WithPayload(&body)
	}

	raw := user.Raw != nil && *user.Raw
	if raw {
		var body models.ObjectCollection
		body.Count = int64(len(vms))
		body.Results = get_raw_virtualmachines(vms)
		return operations.NewVSphereGetVMByMoRefOK().WithPayload(&body)
	}

	return middleware.ResponderFunc(func(rw http.ResponseWriter, p runtime.Producer) {
		write_virtualmachines(vms, rw)
	})
}
