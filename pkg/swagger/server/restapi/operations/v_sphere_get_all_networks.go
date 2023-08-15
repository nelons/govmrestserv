// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// VSphereGetAllNetworksHandlerFunc turns a function with the right signature into a v sphere get all networks handler
type VSphereGetAllNetworksHandlerFunc func(VSphereGetAllNetworksParams) middleware.Responder

// Handle executing the request and returning a response
func (fn VSphereGetAllNetworksHandlerFunc) Handle(params VSphereGetAllNetworksParams) middleware.Responder {
	return fn(params)
}

// VSphereGetAllNetworksHandler interface for that can handle valid v sphere get all networks params
type VSphereGetAllNetworksHandler interface {
	Handle(VSphereGetAllNetworksParams) middleware.Responder
}

// NewVSphereGetAllNetworks creates a new http.Handler for the v sphere get all networks operation
func NewVSphereGetAllNetworks(ctx *middleware.Context, handler VSphereGetAllNetworksHandler) *VSphereGetAllNetworks {
	return &VSphereGetAllNetworks{Context: ctx, Handler: handler}
}

/*
	VSphereGetAllNetworks swagger:route GET /vsphere/{vcenter}/network vSphereGetAllNetworks

Gets a list of networks
*/
type VSphereGetAllNetworks struct {
	Context *middleware.Context
	Handler VSphereGetAllNetworksHandler
}

func (o *VSphereGetAllNetworks) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewVSphereGetAllNetworksParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}