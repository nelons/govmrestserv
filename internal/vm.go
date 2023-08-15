package internal

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/nelons/vsphere-rest-server/pkg/swagger/server/restapi/operations"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
)

/*
Output Functions
*/
func write_virtualmachines(items []mo.VirtualMachine, props []string, rw http.ResponseWriter) {
	rw.WriteHeader(200)

	jsonObj := gabs.New()
	jsonObj.Set(int64(len(items)), "count")
	jsonObj.Array("results")

	for _, item := range items {
		objData, err := serialise_object(item, nil, props)

		if err == nil {
			// Add to the results array
			jsonObj.ArrayAppend(objData, "results")
		}
	}

	out := jsonObj.String()
	rw.Write([]byte(out))
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

	//vms, err := vcenter_getvms_summary(vc.client, ctx)
	var vms []mo.VirtualMachine
	err = vcenter_get_objects(vc.client, ctx, ObjectType_VirtualMachine, []string{"summary"}, &vms)
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

	var props []string
	if user.Props != nil {
		props = strings.Split(*user.Props, ",")
	}

	var vms []mo.VirtualMachine
	err = vcenter_get_object_byname(vc.client, ctx, ObjectType_VirtualMachine, user.Vmname, props, &vms)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetVMByNameBadRequest().WithPayload(&body)
	}

	return middleware.ResponderFunc(func(rw http.ResponseWriter, p runtime.Producer) {
		write_virtualmachines(vms, props, rw)
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

	var props []string
	if user.Props != nil {
		props = strings.Split(*user.Props, ",")
	}

	var vms []mo.VirtualMachine
	err = vcenter_get_object_byref(vc.client, ctx, ObjectType_VirtualMachine, user.Moref, props, &vms)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetVMByMoRefBadRequest().WithPayload(&body)
	}

	return middleware.ResponderFunc(func(rw http.ResponseWriter, p runtime.Producer) {
		write_virtualmachines(vms, props, rw)
	})
}

func post_vsphere_vm_power(user operations.VSphereChangeVMPowerStateParams) middleware.Responder {
	vc, err := validate_request(user.HTTPRequest, user.VRSToken, user.Vcenter)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereChangeVMPowerStateBadRequest().WithPayload(&body)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var vms []mo.VirtualMachine
	err = vcenter_get_object_byref(vc.client, ctx, ObjectType_VirtualMachine, user.Moref, []string{}, &vms)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereChangeVMPowerStateBadRequest().WithPayload(&body)
	}

	if len(vms) == 1 {
		vm := vms[0]

		mobj := object.NewVirtualMachine(vc.client, vm.Self)
		power_state, err := mobj.PowerState(ctx)
		if err == nil {
			fmt.Printf("Current state: %v, Desired State: %v\n", power_state, user.State)

			if user.State == "on" {
				mobj.PowerOn(ctx)

			} else if user.State == "off" {
				mobj.PowerOff(ctx)

			}
		}
	}

	return operations.NewVSphereChangeVMPowerStateOK()
}

// Snapshots ?
