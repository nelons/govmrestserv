// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// VSphereGetAllClustersHandlerFunc turns a function with the right signature into a v sphere get all clusters handler
type VSphereGetAllClustersHandlerFunc func(VSphereGetAllClustersParams) middleware.Responder

// Handle executing the request and returning a response
func (fn VSphereGetAllClustersHandlerFunc) Handle(params VSphereGetAllClustersParams) middleware.Responder {
	return fn(params)
}

// VSphereGetAllClustersHandler interface for that can handle valid v sphere get all clusters params
type VSphereGetAllClustersHandler interface {
	Handle(VSphereGetAllClustersParams) middleware.Responder
}

// NewVSphereGetAllClusters creates a new http.Handler for the v sphere get all clusters operation
func NewVSphereGetAllClusters(ctx *middleware.Context, handler VSphereGetAllClustersHandler) *VSphereGetAllClusters {
	return &VSphereGetAllClusters{Context: ctx, Handler: handler}
}

/*
	VSphereGetAllClusters swagger:route GET /vsphere/{vcenter}/cluster vSphereGetAllClusters

Gets the hosts found at the endpoint.
*/
type VSphereGetAllClusters struct {
	Context *middleware.Context
	Handler VSphereGetAllClustersHandler
}

func (o *VSphereGetAllClusters) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewVSphereGetAllClustersParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}