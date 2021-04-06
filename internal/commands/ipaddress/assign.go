package ipaddress

import (
	"fmt"
	"github.com/UpCloudLtd/cli/internal/commands"
	"github.com/UpCloudLtd/cli/internal/output"
	"github.com/UpCloudLtd/cli/internal/ui"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/spf13/pflag"
)

type assignCommand struct {
	*commands.BaseCommand
	//	ipSvc     service.IpAddress
	//	serverSvc service.Server
	//	req       request.AssignIPAddressRequest
	floating   bool
	access     string
	family     string
	serverUUID string
	mac        string
	zone       string
}

// AssignCommand creates the 'ip-address assign' command
func AssignCommand() commands.Command {
	return &assignCommand{
		BaseCommand: commands.New("assign", "Assign an ip address"),
	}
}

const (
	defaultAccess = upcloud.IPAddressAccessPublic
	defaultFamily = upcloud.IPAddressFamilyIPv4
)

// InitCommand implements Command.InitCommand
func (s *assignCommand) InitCommand() {
	fs := &pflag.FlagSet{}
	fs.StringVar(&s.access, "access", defaultAccess, "Is address for utility or public network.")
	fs.StringVar(&s.family, "family", defaultFamily, "The address family of new IP address.")
	fs.StringVar(&s.serverUUID, "server", "", "The server the ip address is assigned to.")
	fs.StringVar(&s.mac, "mac", "", "MAC address of server interface to assign address to. Required for non-floating addresses.")
	fs.StringVar(&s.zone, "zone", "", "Zone of address, required when assigning a detached floating IP address.")
	fs.BoolVar(&s.floating, "floating", false, "Whether the address to be assigned is a floating one.")
	s.AddFlags(fs)
}

// Execute implements NewCommand.Execute
func (s *assignCommand) Execute(exec commands.Executor, arg string) (output.Output, error) {
	if s.floating && s.zone == "" && s.mac == "" {
		return nil, fmt.Errorf("MAC or zone is required for floating IP")
	}
	if !s.floating && s.serverUUID == "" {
		return nil, fmt.Errorf("server is required for non-floating IP")
	}

	if s.serverUUID != "" {
		_, err := exec.Server().GetServerDetails(&request.GetServerDetailsRequest{UUID: s.serverUUID})
		if err != nil {
			return nil, err
		}
	}
	target := s.mac
	if target == "" {
		target = s.zone
	}
	msg := fmt.Sprintf("Assigning IP Address to %v", target)
	logline := exec.NewLogEntry(msg)
	logline.StartedNow()

	res, err := exec.IPAddress().AssignIPAddress(&request.AssignIPAddressRequest{
		Access:     s.access,
		Family:     s.family,
		ServerUUID: s.serverUUID,
		Floating:   upcloud.FromBool(s.floating),
		MAC:        s.mac,
		Zone:       s.zone,
	})
	if err != nil {
		logline.SetMessage(ui.LiveLogEntryErrorColours.Sprintf("%s: failed (%v)", msg, err.Error()))
		logline.SetDetails(err.Error(), "error: ")
		return nil, err
	}
	logline.SetMessage(fmt.Sprintf("%s: success", msg))
	logline.MarkDone()
	return output.Marshaled{Value: res}, nil
}
