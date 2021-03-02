package ipaddress

import (
	"fmt"
	"github.com/UpCloudLtd/cli/internal/commands"
	"github.com/UpCloudLtd/cli/internal/ui"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/pflag"
	"io"
)

// ListCommand creates the "ip-address list" command
func ListCommand(service service.IpAddress) commands.Command {
	return &listCommand{
		BaseCommand: commands.New("list", "List ip addresses"),
		service:     service,
	}
}

type listCommand struct {
	*commands.BaseCommand
	service        service.IpAddress
	header         table.Row
	columnKeys     []string
	visibleColumns []string
}

// InitCommand implements Command.MakeExecuteCommand
func (s *listCommand) InitCommand() {
	s.header = table.Row{"Address", "Access", "Family", "Part of Plan", "PTR Record", "Server", "MAC", "Floating", "Zone"}
	s.columnKeys = []string{"address", "access", "family", "partofplan", "ptrrecord", "server", "mac", "floating", "zone"}
	s.visibleColumns = []string{"address", "access", "family", "partofplan", "ptrrecord", "server", "mac", "floating", "zone"}
	flags := &pflag.FlagSet{}
	s.AddVisibleColumnsFlag(flags, &s.visibleColumns, s.columnKeys, s.visibleColumns)
	s.AddFlags(flags)
}

// MakeExecuteCommand implements Command.MakeExecuteCommand
func (s *listCommand) MakeExecuteCommand() func(args []string) (interface{}, error) {
	return func(args []string) (interface{}, error) {
		ips, err := s.service.GetIPAddresses()
		if err != nil {
			return nil, err
		}
		return ips, nil
	}
}

// HandleOutput implements Command.HandleOutput
func (s *listCommand) HandleOutput(writer io.Writer, out interface{}) error {
	ips := out.(*upcloud.IPAddresses)

	t := ui.NewDataTable(s.columnKeys...)
	t.OverrideColumnKeys(s.visibleColumns...)
	t.SetHeader(s.header)

	for _, ip := range ips.IPAddresses {
		t.Append(table.Row{
			ui.DefaultAddressColours.Sprint(ip.Address),
			ip.Access,
			ip.Family,
			ui.FormatBool(ip.PartOfPlan.Bool()),
			ip.PTRRecord,
			ui.DefaultUUUIDColours.Sprint(ip.ServerUUID),
			ip.MAC,
			ui.FormatBool(ip.Floating.Bool()),
			ip.Zone})
	}

	_, _ = fmt.Fprintln(writer, t.Render())
	return nil
}
