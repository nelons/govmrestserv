// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/nelons/vsphere-rest-server/pkg/swagger/server/models"
)

// VSphereGetAllHostsSummaryOKCode is the HTTP code returned for type VSphereGetAllHostsSummaryOK
const VSphereGetAllHostsSummaryOKCode int = 200

/*
VSphereGetAllHostsSummaryOK Successful request. Returns JSON containing a count and the collection of objects.

swagger:response vSphereGetAllHostsSummaryOK
*/
type VSphereGetAllHostsSummaryOK struct {

	/*
	  In: Body
	*/
	Payload *models.ObjectCollection `json:"body,omitempty"`
}

// NewVSphereGetAllHostsSummaryOK creates VSphereGetAllHostsSummaryOK with default headers values
func NewVSphereGetAllHostsSummaryOK() *VSphereGetAllHostsSummaryOK {

	return &VSphereGetAllHostsSummaryOK{}
}

// WithPayload adds the payload to the v sphere get all hosts summary o k response
func (o *VSphereGetAllHostsSummaryOK) WithPayload(payload *models.ObjectCollection) *VSphereGetAllHostsSummaryOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the v sphere get all hosts summary o k response
func (o *VSphereGetAllHostsSummaryOK) SetPayload(payload *models.ObjectCollection) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *VSphereGetAllHostsSummaryOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// VSphereGetAllHostsSummaryBadRequestCode is the HTTP code returned for type VSphereGetAllHostsSummaryBadRequest
const VSphereGetAllHostsSummaryBadRequestCode int = 400

/*
VSphereGetAllHostsSummaryBadRequest A general failure occured, more details are contained within the message.

swagger:response vSphereGetAllHostsSummaryBadRequest
*/
type VSphereGetAllHostsSummaryBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.BadRequestError `json:"body,omitempty"`
}

// NewVSphereGetAllHostsSummaryBadRequest creates VSphereGetAllHostsSummaryBadRequest with default headers values
func NewVSphereGetAllHostsSummaryBadRequest() *VSphereGetAllHostsSummaryBadRequest {

	return &VSphereGetAllHostsSummaryBadRequest{}
}

// WithPayload adds the payload to the v sphere get all hosts summary bad request response
func (o *VSphereGetAllHostsSummaryBadRequest) WithPayload(payload *models.BadRequestError) *VSphereGetAllHostsSummaryBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the v sphere get all hosts summary bad request response
func (o *VSphereGetAllHostsSummaryBadRequest) SetPayload(payload *models.BadRequestError) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *VSphereGetAllHostsSummaryBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}