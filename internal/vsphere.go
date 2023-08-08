package internal

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"

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

	info := client.ServiceContent.About
	fmt.Printf("Connected to vCenter version %s\n", info.Version)
	return client, nil
}

func vcenter_getvms_summary(client *vim25.Client, ctx context.Context) ([]mo.VirtualMachine, error) {
	m := view.NewManager(client)

	v, err := m.CreateContainerView(ctx, client.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return nil, err
	}
	defer v.Destroy(ctx)

	var vms []mo.VirtualMachine
	err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary"}, &vms)
	if err != nil {
		return vms, err
	}

	return vms, nil
}

func vcenter_getvm_byname(client *vim25.Client, ctx context.Context, vm_name string) ([]mo.VirtualMachine, error) {
	m := view.NewManager(client)

	v, err := m.CreateContainerView(ctx, client.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return nil, err
	}

	defer v.Destroy(ctx)

	var vms []mo.VirtualMachine
	err = v.RetrieveWithFilter(ctx, []string{"VirtualMachine"}, []string{}, &vms, property.Filter{"name": vm_name})
	if err != nil {
		return nil, err
	}

	return vms, nil
}

func vcenter_getvm_bymoref(client *vim25.Client, ctx context.Context, vm_moref string) ([]mo.VirtualMachine, error) {
	m := view.NewManager(client)

	v, err := m.CreateContainerView(ctx, client.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return nil, err
	}

	defer v.Destroy(ctx)

	var moref types.ManagedObjectReference
	moref.Type = "VirtualMachine"
	moref.Value = vm_moref

	var vms []mo.VirtualMachine
	err = v.RetrieveWithFilter(ctx, []string{"VirtualMachine"}, []string{}, &vms, property.Filter{"Self": moref})
	if err != nil {
		return nil, err
	}

	return vms, nil
}
