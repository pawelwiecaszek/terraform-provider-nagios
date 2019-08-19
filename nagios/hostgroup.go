package nagios

// NewHostgroup initiates the HTTP POST to the Nagios API to create a hostgroup
func (c *Client) NewHostgroup(hostgroup *Hostgroup) error {
	configURL := "/config/hostgroup"

	_, err := c.post(configURL, &hostgroup)

	if err != nil {
		return err
	}

	// func (c *Client) post(configURL string, requestBody interface{}) ([]byte, error) {

	return nil
}

func (c *Client) GetHostgroup(name string) (*Hostgroup, error) {
	var hostgroup Hostgroup

	objectURL := "/objects/hostgroup"

	err := c.get(objectURL, &hostgroup)

	if err != nil {
		return nil, err
	}

	return &hostgroup, nil
}
