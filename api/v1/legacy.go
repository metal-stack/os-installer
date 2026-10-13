package v1

const (
	LegacyInstallPath = "/etc/metal/install.yaml"
)

type (

	// InstallerConfig contains legacy configuration items which are
	// used to install the os.
	// It must be serialized to /etc/metal/install.yaml to guarantee compatibility for older
	// firewall-controller and lldpd
	InstallerConfig struct {
		// Hostname of the machine
		Hostname string `yaml:"hostname"`
		// Networks all networks connected to this machine
		Networks []*V1MachineNetwork `yaml:"networks"`
		// MachineUUID is the unique UUID for this machine, usually the board serial.
		MachineUUID string `yaml:"machineuuid"`
		// SSHPublicKey of the user
		SSHPublicKey string `yaml:"sshpublickey"`
		// Password is the password for the metal user.
		Password string `yaml:"password"`
		// Console specifies where the kernel should connect its console to.
		Console string `yaml:"console"`
		// Timestamp is the the timestamp of installer config creation.
		Timestamp string `yaml:"timestamp"`
		// Nics are the network interfaces of this machine including their neighbors.
		Nics []*V1MachineNic `yaml:"nics"`
		// VPN is the config for connecting machine to VPN
		VPN *V1MachineVPN `yaml:"vpn"`
		// Role is either firewall or machine
		Role string `yaml:"role"`
		// RaidEnabled is set to true if any raid devices are specified
		RaidEnabled bool `yaml:"raidenabled"`
		// RootUUID is the fs uuid if the root fs
		RootUUID string `yaml:"root_uuid"`
		// FirewallRules if not empty firewall rules to enforce
		FirewallRules *V1FirewallRules `yaml:"firewall_rules"`
		// DNSServers for the machine
		DNSServers []*V1DNSServer `yaml:"dns_servers"`
		// NTPServers for the machine
		NTPServers []*V1NTPServer `yaml:"ntp_servers"`
	}

	// Copies of metal-go models.V1* structs in use in Installerconfig
	// to prevent the import of metal-go.

	V1MachineNetwork struct {
		// ASN number for this network in the bgp configuration
		// Required: true
		Asn *int64 `json:"asn" yaml:"asn"`
		// the destination prefixes of this network
		// Required: true
		Destinationprefixes []string `json:"destinationprefixes" yaml:"destinationprefixes"`
		// the ip addresses of the allocated machine in this vrf
		// Required: true
		Ips []string `json:"ips" yaml:"ips"`
		// if set to true, packets leaving this network get masqueraded behind interface ip
		// Required: true
		Nat *bool `json:"nat" yaml:"nat"`
		// nattypev2
		// Required: true
		Nattypev2 *string `json:"nattypev2" yaml:"nattypev2"`
		// the networkID of the allocated machine in this vrf
		// Required: true
		Networkid *string `json:"networkid" yaml:"networkid"`
		// the network type, types can be looked up in the network package of metal-lib
		// Required: true
		Networktype *string `json:"networktype" yaml:"networktype"`
		// networktypev2
		// Required: true
		Networktypev2 *string `json:"networktypev2" yaml:"networktypev2"`
		// the prefixes of this network
		// Required: true
		Prefixes []string `json:"prefixes" yaml:"prefixes"`
		// indicates whether this network is the private network of this machine
		// Required: true
		Private *bool `json:"private" yaml:"private"`
		// project of this network, empty string if not project scoped
		// Required: true
		Projectid *string `json:"projectid" yaml:"projectid"`
		// if set to true, this network can be used for underlay communication
		// Required: true
		Underlay *bool `json:"underlay" yaml:"underlay"`
		// the vrf of the allocated machine
		// Required: true
		Vrf *int64 `json:"vrf" yaml:"vrf"`
	}

	V1MachineNic struct {
		// the unique identifier of this network interface
		// Required: true
		Identifier *string `json:"identifier" yaml:"identifier"`
		// the mac address of this network interface
		// Required: true
		Mac *string `json:"mac" yaml:"mac"`
		// the name of this network interface
		// Required: true
		Name *string `json:"name" yaml:"name"`
		// the neighbors visible to this network interface
		// Required: true
		Neighbors []*V1MachineNic `json:"neighbors" yaml:"neighbors"`
	}

	V1MachineVPN struct {
		// address of VPN control plane
		// Required: true
		Address *string `json:"address" yaml:"address"`
		// auth key used to connect to VPN
		// Required: true
		AuthKey *string `json:"auth_key" yaml:"auth_key"`
		// connected to the VPN
		// Required: true
		Connected *bool `json:"connected" yaml:"connected"`
	}

	V1FirewallRules struct {
		// list of egress rules to be deployed during firewall allocation
		Egress []*V1FirewallEgressRule `json:"egress" yaml:"egress"`
		// list of ingress rules to be deployed during firewall allocation
		Ingress []*V1FirewallIngressRule `json:"ingress" yaml:"ingress"`
	}

	V1FirewallEgressRule struct {
		// an optional comment describing what this rule is used for
		Comment string `json:"comment,omitempty" yaml:"comment,omitempty"`
		// the ports affected by this rule
		// Required: true
		Ports []int32 `json:"ports" yaml:"ports"`
		// the protocol for the rule, defaults to tcp
		// Enum: ["tcp","udp"]
		Protocol string `json:"protocol,omitempty" yaml:"protocol,omitempty"`
		// the cidrs affected by this rule
		// Required: true
		To []string `json:"to" yaml:"to"`
	}

	V1FirewallIngressRule struct {
		// an optional comment describing what this rule is used for
		Comment string `json:"comment,omitempty" yaml:"comment,omitempty"`
		// the cidrs affected by this rule
		// Required: true
		From []string `json:"from" yaml:"from"`
		// the ports affected by this rule
		// Required: true
		Ports []int32 `json:"ports" yaml:"ports"`
		// the protocol for the rule, defaults to tcp
		// Enum: ["tcp","udp"]
		Protocol string `json:"protocol,omitempty" yaml:"protocol,omitempty"`
		// the cidrs affected by this rule
		To []string `json:"to" yaml:"to"`
	}

	V1DNSServer struct {
		// ip address of this dns server
		// Required: true
		IP *string `json:"ip" yaml:"ip"`
	}

	V1NTPServer struct {
		// ip address or dns hostname of this ntp server
		// Required: true
		Address *string `json:"address" yaml:"address"`
	}
)
