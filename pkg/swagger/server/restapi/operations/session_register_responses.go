// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"
)

// SessionRegisterOKCode is the HTTP code returned for type SessionRegisterOK
const SessionRegisterOKCode int = 200

/*
SessionRegisterOK Registers the user session

swagger:response sessionRegisterOK
*/
type SessionRegisterOK struct {

	/*
	  In: Body
	*/
	Payload *SessionRegisterOKBody `json:"body,omitempty"`
}

// NewSessionRegisterOK creates SessionRegisterOK with default headers values
func NewSessionRegisterOK() *SessionRegisterOK {

	return &SessionRegisterOK{}
}

// WithPayload adds the payload to the session register o k response
func (o *SessionRegisterOK) WithPayload(payload *SessionRegisterOKBody) *SessionRegisterOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the session register o k response
func (o *SessionRegisterOK) SetPayload(payload *SessionRegisterOKBody) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *SessionRegisterOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// SessionRegisterBadRequestCode is the HTTP code returned for type SessionRegisterBadRequest
const SessionRegisterBadRequestCode int = 400

/*
SessionRegisterBadRequest Session already exists.

swagger:response sessionRegisterBadRequest
*/
type SessionRegisterBadRequest struct {
}

// NewSessionRegisterBadRequest creates SessionRegisterBadRequest with default headers values
func NewSessionRegisterBadRequest() *SessionRegisterBadRequest {

	return &SessionRegisterBadRequest{}
}

// WriteResponse to the client
func (o *SessionRegisterBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(400)
}

// SessionRegisterInternalServerErrorCode is the HTTP code returned for type SessionRegisterInternalServerError
const SessionRegisterInternalServerErrorCode int = 500

/*
SessionRegisterInternalServerError Internal server error

swagger:response sessionRegisterInternalServerError
*/
type SessionRegisterInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *SessionRegisterInternalServerErrorBody `json:"body,omitempty"`
}

// NewSessionRegisterInternalServerError creates SessionRegisterInternalServerError with default headers values
func NewSessionRegisterInternalServerError() *SessionRegisterInternalServerError {

	return &SessionRegisterInternalServerError{}
}

// WithPayload adds the payload to the session register internal server error response
func (o *SessionRegisterInternalServerError) WithPayload(payload *SessionRegisterInternalServerErrorBody) *SessionRegisterInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the session register internal server error response
func (o *SessionRegisterInternalServerError) SetPayload(payload *SessionRegisterInternalServerErrorBody) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *SessionRegisterInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
