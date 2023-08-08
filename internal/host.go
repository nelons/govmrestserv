package internal

import (
	"context"

	"github.com/go-openapi/runtime/middleware"
	"github.com/nelons/vsphere-rest-server/pkg/swagger/server/restapi/operations"
)

func get_vsphere_get_host(user operations.VSphereGetAllHostsSummaryParams) middleware.Responder {
	_, err := validate_request(user.HTTPRequest, user.VRSToken, user.Vcenter)
	if err != nil {
		body := create_badrequesterror(err.Error())
		return operations.NewVSphereGetVMByMoRefBadRequest().WithPayload(&body)
	}

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// TODO: vsphere get hosts

	return operations.NewVSphereGetAllHostsSummaryBadRequest()
}
