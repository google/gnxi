## Test classes

### Module get

 *  `get.Get`

    Tests that gNMI Get of an xpath returns no error.

    Args:
     *  `xpath`: gNMI path to read, like "system/config"

 *  `get.GetCompare`

    Compares that gNMI Get of an xpath returns the expected value.

    Args:
     *  `xpath`: gNMI path to read, like "system/config"
     *  `want`: Expected value; can be numeric, string or JSON-IETF
     *  `retries`: Optional. Number of retries if the assertion fails
     *  `retry_delay`: Optional. Delay, in seconds, between retries. Default 10

 *  `get.GetJsonCheck`

    Tests that gNMI Get of an xpath returns a schema-valid JSON response.

    Args:
     *  `xpath`: gNMI path to read, like "system/config"
     *  `model`: Python binding class to check the JSON reply against.
        like `system.config.config`.
     *  `retries`: Optional. Number of retries if the assertion fails.
     *  `retry_delay`: Optional. Delay, in seconds, between retries. Default 10.

        The binding classes are in the `oc_config_validate.models` package.

 *  `get.GetJsonCheckCompare`

    Checks for schema validity and compares a gNMI Get response.

    Args:
     *  `xpath`: gNMI path to read, like "system/config"
     *  `model`: Python binding class to check the JSON reply against.
        like `system.config.config`. 
     *  `want_json`: Expected JSON-IETF value.
     *  `retries`: Optional. Number of retries if the assertion fails.
     *  `retry_delay`: Optional. Delay, in seconds, between retries. Default 10.

        The binding classes are in the `oc_config_validate.models` package.

### Module set

 *  `set.SetUpdate`

    Sends gNMI Set Update of an xpath with a value.

    Args:
     *  `xpath`: gNMI path to update, like "system/config"
     *  `value`: Value to set; can be numeric, string or JSON-IETF.

 *  `set.SetDelete`

    Sends gNMI Set Delete of an xpath.

    Args:
     *  `xpath`: gNMI path to delete, like "system/config"
     
 *  `set.JsonCheckSetUpdate`

    Sends gNMI Set with a schema-checked JSON-IETF value.

    Args:
     *  `xpath`: gNMI path to read, like "system/config"
     *  `json_value`: JSON-IETF value to check and set.
     *  `model`: Python binding class to check the JSON reply against.
        like `system.config.config`.

### Module setget

 *  `setget.JsonCheck`

    Sends gNMI Set and later Get, schema-checking the JSON-IETF value.

     1.  The intended JSON-IETF configuration is checked for schema validity.
     1.  It is sent in a gNMI Set request.
     1.  The same path is used for a gNMI Get request.
     1.  The returned value is checked for schema validity

    Args:
     *  `xpath`: gNMI path to write and read, like "system/config"
     *  `json_value`: JSON-IETF value to check set and get.
     *  `model`: Python binding class to check the JSON reply against.
        like `system.config.config`.
     *  `retries`: Optional. Number of retries if the assertion fails.
     *  `retry_delay`: Optional. Delay, in seconds, between retries. Default 10.

 *  `setget.JsonCheckCompare`

    Does what setget.JsonCheck does, but also compares the JSON Get reply.

    1. The intended JSON-IETF configuration is checked for schema validity.
    1. It is sent in a gNMI Set request.
    1. The same path is used for a gNMI Get request.
    1. The returned value is checked for schema validity
    1. The returned value is compared with the sent value

    Args:
     *  `xpath`: gNMI path to write and read, like "system/config"
     *  `json_value`: JSON-IETF value to check set, get and compare.
     *  `model`: Python binding class to check the JSON reply against.
        like `system.config.config`.
     *  `retries`: Optional. Number of retries if the assertion fails.
     *  `retry_delay`: Optional. Delay, in seconds, between retries. Default 10.

### Module config_state

 *  `config_state.SetConfigCheckState`
 
    Configures on the /config container and checks the /state container.
    
        1. The intended JSON-IETF configuration is checked for schema validity.
        1. It is sent in a gNMI Set request,  to the /config container.
        1. The same container is fetched in a gNMI Get request and checked for
             schema validity. It is compared with the sent configuration.
        1. The /state container is fetched in a gNMI Get request and checked for
             schema validity. It is compared with the sent configuration.

    It will retry the last /state container up to 10 times, in case the device
      needs some time to update the state information.

    Args:
     *  `xpath`: gNMI path to write and read, without ending /config or /state.
     *  `json_value`: JSON-IETF value to check set, get and compare.
     *  `model`: Python binding module to check the against. 
        Do not specify a Python class, the test will instantiance the config()
        and state() objects from the module.
     *  `retries`: Optional. Number of retries if the assertion fails.
     *  `retry_delay`: Optional. Delay, in seconds, between retries. Default 10.

 *  `config_state.DeleteConfigCheckState`
 
    Deletes the xpath and checks the /config and /state container are no longer
    there.

    1. gNMI Get request validates that the /config container exists.
    1. gNMI Delete request removes the xpath.
    1. gNMI Get request validates that the /config container no longer exists.
    1. gNMI Get request validates that the /state container no longer exists.

    All arguments are read from the Test YAML description.

    Args:
     *  `xpath`: gNMI path to delete, without ending /config or /state.
     *  `retries`: Optional. Number of retries if the assertion fails.
     *  `retry_delay`: Optional. Delay, in seconds, between retries. Default 10.

### Module static_route

By default, the tests do 3 retries, with 10 seconds delay, if the assertion fails.

 *  `static_route.AddStaticRoute`

    Tests configuring a static route.

    1. A gNMI Set message is sent to configure the route.
    1. A gNMI Get message on the /config and /state containers validates it.

    All arguments are read from the Test YAML description.

    Args:
     * `prefix`: Destination prefix of the static route.
     * `next_hop`: IP of the next hop of the route.
     * `index`: Index of the next hop for the prefix. Defaults to 0.
     * `metric`: Optional numeric metric of the next hop for the prefix.
     * `description`: Optional text description of the route.

 *  `static_route.RemoveStaticRoute`
 
    Tests removing a static route.

    1. gNMI Get message on the /config container, to check it is configured.
    1. gNMI Set message to delete the route.
    1. gNMI Get message on the /config container to check it is not there.

    All arguments are read from the Test YAML description.

    Args:
     * `prefix`: Destination prefix of the static route.
     * `index`: Index of the next hop for the prefix. Defaults to 0.

 *  `static_route.CheckRouteState`
 
     Checks the state on a static route.

    1. A gNMI Get message on the /state container.

    All arguments are read from the Test YAML description.

    Args:
     * `prefix`: Destination prefix of the static route.
     * `next_hop`: IP of the next hop of the route.
     * `index`: Index of the next hop for the prefix. Defaults to 0.
     * `metric`: Optional numeric metric of the next hop for the prefix.
     * `description`: Optional text description of the route.
 
 *  `static_route.CheckRouteConfig`
 
     Checks the config on a static route.

    1. A gNMI Get message on the /config container.

    All arguments are read from the Test YAML description.

    Args:
     * `prefix`: Destination prefix of the static route.
     * `next_hop`: IP of the next hop of the route.
     * `index`: Index of the next hop for the prefix. Defaults to 0.
     * `metric`: Optional numeric metric of the next hop for the prefix.
     * `description`: Optional text description of the route.

### Module subif_ip

By default, the tests do 5 retries, with 15 seconds delay, if the assertion fails.

 *  `subif_ip.SetSubifDhcp`
 
    Tests configuring DHCP on a subinterface.

    1. A gNMI Set message is sent to configure the subinterface.
    1. A gNMI Get message on the /config container validates it.

    All arguments are read from the Test YAML description.

    Args:
     * `interface`: Name of the physical interface.
     * `index`: Index of the subinterface, defaults to 0.
     * `dhcp`: True to enable DHCP, defaults to False.
     
 *  `subif_ip.CheckSubifDhcpState`
 
    Checks the DHCP state on a subinterface.

    1. A gNMI Get message on the /state container.
    
    All arguments are read from the Test YAML description.

    Args:
     * `interface`: Name of the physical interface.
     * `index`: Index of the subinterface, defaults to 0.
     * `dhcp`: True to enable DHCP, defaults to False.

 *  `subif_ip.CheckSubifDhcpConfig`
 
    Checks the DHCP config on a subinterface.

    1. A gNMI Get message on the /config container.
    
    All arguments are read from the Test YAML description.

    Args:
     * `interface`: Name of the physical interface.
     * `index`: Index of the subinterface, defaults to 0.
     * `dhcp`: True to enable DHCP, defaults to False.

 *  `subif_ip.AddSubifIp`
   
    Tests configuring an IP on a subinterface.

    1. A gNMI Set message is sent to configure the subinterface.
    1. A gNMI Get message on the /config container validates it.

    All arguments are read from the Test YAML description.

    Args:
     * `interface`: Name of the physical interface.
     * `index`: Index of the subinterface, defaults to 0.
     * `address`: IPv4 address to add.
     * `prefix_length`: Prefix lenght of the IPv4 address to add.

 *  `subif_ip.RemoveSubifIp`
   
    Tests removing an IP on a subinterface.

    1. gNMI Get message on the /config container, to check it is configured.
    1. gNMI Set message to delete the ip.
    1. gNMI Get message on the /config container to check it is not there.
    1. gNMI Get message on the /state container to check it is not there.

    All arguments are read from the Test YAML description.

    Args:
     * `interface`: Name of the physical interface.
     * `index`: Index of the subinterface, defaults to 0.
     * `address`: IPv4 address.
     * `prefix_length`: Prefix lenght of the IPv4 address.

 *  `subif_ip.CheckSubifIpState`

    Checks the state on an ip address configured on a subinterface.

    1. A gNMI Get message on the /state container.

    All arguments are read from the Test YAML description.

    Args:
     * `interface`: Name of the physical interface.
     * `index`: Index of the subinterface, defaults to 0.
     * `address`: IPv4 address.
     * `prefix_length`: Prefix lenght of the IPv4 address.
     
 *  `subif_ip.CheckSubifIpConfig`

    Checks the configuration on an ip address on a subinterface.

    1. A gNMI Get message on the /config container.

    All arguments are read from the Test YAML description.

    Args:
     * `interface`: Name of the physical interface.
     * `index`: Index of the subinterface, defaults to 0.
     * `address`: IPv4 address.
     * `prefix_length`: Prefix lenght of the IPv4 address.
