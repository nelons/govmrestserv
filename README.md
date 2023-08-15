# vSphere Rest Server

## Disclaimer

I'm not a golang dev, so expect errors, bugs, and definitely nothing adhering to best practices (so if you do find the latter then it's just pure luck).

## Introduction

This application is a web server that allows you to interact with multiple vSphere Servers via REST API returning data structures that match the SOAP interfaces.

There is an official vSphere REST API implementation, and documentation is at https://developer.vmware.com/apis/vsphere-automation/latest/vcenter/

This can be used against vcsim images/containers.

## Building

As I'm a Windows dev, this has been built by default to run on Windows, but I do want to make it as OS-agnostic as possible. Still, there might be oversights.

make build.windows

## Running

The application supports the following command line parameters:

-help -> displays the command line options  
-test "vc sdk url" "username" "password" -> connects to the URL and prints out information about what can be found. This does not start the web server.  
-app -> runs the application from the command line.  
-service install -> installs the application as a service.  
-service uninstall -> removes the service.  
-service start -> starts the service  
-service stop -> stops.  

If not command line parameters are provided, or "-app" then the application will start a blocking web server, Ctrl-C or equivalent to exit.

## Configuration

Create/Edit "config.json" in the same folder with the options:

port - integer representing the port to listen on
certificate_file - file containing the certificate to use
certificate_key - file containing the private key for the above certificate

If the latter two are not provided or found, then a self-signed certificate will be generated to be used by the web server.

## API

The included file doc/index.html documents all the API calls.

### Some examples

POST /session/register to create a session with a secret string. This returns a token, which should be included in all future headers (with the key "VRS-Token")

POST /vsphere/connect will connect to a provided vCenter Server. You can provide a nice name for the connection, which will be needed in future calls to identify the target server you want to run commands against.

GET /vpshere/{name}/vm/{name} to get an extract of a virtual machine by name.

GET /vsphere/{name}/vm/{moref} as above, but identifying the VM by managedreference.

## Output

Each SOAP data structure has an additional member "_typename", which can be used to map to a data structure detailed in the vim WSDL.

## Defects/Shortcomings

There are no limits on the amount of entities returned, or paging of results.
Logging is non-existant at this point.

## Why

Great question :)

I wrote this because the C# SOAP implementation was deprecated and I couldn't find a way to generate the service code with Visual Studio/tools. 

I also wanted to be able to query ESXi as well as vCenter (the REST interface was only available against vCenter), and get the full range of data returned by the SOAP interfaces.