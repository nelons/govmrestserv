// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"context"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// SessionRegisterHandlerFunc turns a function with the right signature into a session register handler
type SessionRegisterHandlerFunc func(SessionRegisterParams) middleware.Responder

// Handle executing the request and returning a response
func (fn SessionRegisterHandlerFunc) Handle(params SessionRegisterParams) middleware.Responder {
	return fn(params)
}

// SessionRegisterHandler interface for that can handle valid session register params
type SessionRegisterHandler interface {
	Handle(SessionRegisterParams) middleware.Responder
}

// NewSessionRegister creates a new http.Handler for the session register operation
func NewSessionRegister(ctx *middleware.Context, handler SessionRegisterHandler) *SessionRegister {
	return &SessionRegister{Context: ctx, Handler: handler}
}

/*
	SessionRegister swagger:route POST /session/register sessionRegister

Register a new Session
*/
type SessionRegister struct {
	Context *middleware.Context
	Handler SessionRegisterHandler
}

func (o *SessionRegister) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewSessionRegisterParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}

// SessionRegisterBody session register body
//
// swagger:model SessionRegisterBody
type SessionRegisterBody struct {

	// secret
	// Required: true
	Secret *string `json:"secret"`
}

// Validate validates this session register body
func (o *SessionRegisterBody) Validate(formats strfmt.Registry) error {
	var res []error

	if err := o.validateSecret(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *SessionRegisterBody) validateSecret(formats strfmt.Registry) error {

	if err := validate.Required("requestBody"+"."+"secret", "body", o.Secret); err != nil {
		return err
	}

	return nil
}

// ContextValidate validates this session register body based on context it is used
func (o *SessionRegisterBody) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (o *SessionRegisterBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *SessionRegisterBody) UnmarshalBinary(b []byte) error {
	var res SessionRegisterBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}

// SessionRegisterInternalServerErrorBody session register internal server error body
//
// swagger:model SessionRegisterInternalServerErrorBody
type SessionRegisterInternalServerErrorBody struct {

	// The error message for failure.
	Error string `json:"error,omitempty"`
}

// Validate validates this session register internal server error body
func (o *SessionRegisterInternalServerErrorBody) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this session register internal server error body based on context it is used
func (o *SessionRegisterInternalServerErrorBody) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (o *SessionRegisterInternalServerErrorBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *SessionRegisterInternalServerErrorBody) UnmarshalBinary(b []byte) error {
	var res SessionRegisterInternalServerErrorBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}

// SessionRegisterOKBody session register o k body
//
// swagger:model SessionRegisterOKBody
type SessionRegisterOKBody struct {

	// The token the user should provide for future requests.
	Token string `json:"token,omitempty"`
}

// Validate validates this session register o k body
func (o *SessionRegisterOKBody) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this session register o k body based on context it is used
func (o *SessionRegisterOKBody) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (o *SessionRegisterOKBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *SessionRegisterOKBody) UnmarshalBinary(b []byte) error {
	var res SessionRegisterOKBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}