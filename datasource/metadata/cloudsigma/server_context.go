package cloudsigma

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"

	"github.com/coreos/coreos-cloudinit/third_party/github.com/cloudsigma/cepgo"
)

const (
	userDataFieldName = "cloudinit-user-data"
)

type serverContextService struct {
	client interface {
		All() (interface{}, error)
		Key(string) (interface{}, error)
		Meta() (map[string]string, error)
		FetchRaw(string) ([]byte, error)
	}
}

func NewServerContextService() *serverContextService {
	return &serverContextService{
		client: cepgo.NewCepgo(),
	}
}

func (_ *serverContextService) IsAvailable() bool {
	productNameFile, err := os.Open("/sys/class/dmi/id/product_name")
	if err != nil {
		return false
	}
	productName := make([]byte, 10)
	_, err = productNameFile.Read(productName)
	return err == nil && string(productName) == "CloudSigma"
}

func (_ *serverContextService) AvailabilityChanges() bool {
	return true
}

func (_ *serverContextService) ConfigRoot() string {
	return ""
}

func (_ *serverContextService) Type() string {
	return "server-context"
}

func (scs *serverContextService) FetchMetadata() ([]byte, error) {
	var (
		inputMetadata struct {
			Name string            `json:"name"`
			UUID string            `json:"uuid"`
			Meta map[string]string `json:"meta"`
			Nics []struct {
				Runtime struct {
					InterfaceType string `json:"interface_type"`
					IPv4          struct {
						IP string `json:"uuid"`
					} `json:"ip_v4"`
				} `json:"runtime"`
			} `json:"nics"`
		}
		outputMetadata struct {
			Hostname   string            `json:"name"`
			PublicKeys map[string]string `json:"public_keys"`
			LocalIPv4  string            `json:"local-ipv4"`
			PublicIPv4 string            `json:"public-ipv4"`
		}
	)

	rawMetadata, err := scs.client.FetchRaw("")
	if err != nil {
		return []byte{}, err
	}

	err = json.Unmarshal(rawMetadata, &inputMetadata)
	if err != nil {
		return []byte{}, err
	}

	if inputMetadata.Name != "" {
		outputMetadata.Hostname = inputMetadata.Name
	} else {
		outputMetadata.Hostname = inputMetadata.UUID
	}

	if key, ok := inputMetadata.Meta["ssh_public_key"]; ok {
		splitted := strings.Split(key, " ")
		outputMetadata.PublicKeys = make(map[string]string)
		outputMetadata.PublicKeys[splitted[len(splitted)-1]] = key
	}

	for _, nic := range inputMetadata.Nics {
		if nic.Runtime.IPv4.IP != "" {
			if nic.Runtime.InterfaceType == "public" {
				outputMetadata.PublicIPv4 = nic.Runtime.IPv4.IP
			} else {
				outputMetadata.LocalIPv4 = nic.Runtime.IPv4.IP
			}
		}
	}

	return json.Marshal(outputMetadata)
}

func (scs *serverContextService) FetchUserdata() ([]byte, error) {
	metadata, err := scs.client.Meta()
	if err != nil {
		return []byte{}, err
	}

	userData, ok := metadata[userDataFieldName]
	if ok && isBase64Encoded(userDataFieldName, metadata) {
		if decodedUserData, err := base64.StdEncoding.DecodeString(userData); err == nil {
			return decodedUserData, nil
		} else {
			return []byte{}, nil
		}
	}

	return []byte(userData), nil
}

func (scs *serverContextService) FetchNetworkConfig(a string) ([]byte, error) {
	return nil, nil
}

func isBase64Encoded(field string, userdata map[string]string) bool {
	base64Fields, ok := userdata["base64_fields"]
	if !ok {
		return false
	}

	for _, base64Field := range strings.Split(base64Fields, ",") {
		if field == base64Field {
			return true
		}
	}
	return false
}
