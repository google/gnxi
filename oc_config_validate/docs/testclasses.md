## Test classes

### Module get

 *  `get.Get`

    Tests that gNMI Get of an xpath returns no error.

    Args:
     *  `xpath`: gNMI path to read.


 *  `get.GetCompare`

    Compares that gNMI Get of an xpath returns the expected value.

    Args:
     *  `xpath`: gNMI path to read.
     *  `want`: Expected value; can be numeric, string or JSON-IETF
     *  `retries`: Optional. Number of retries if the assertion fails
     *  `retry_delay`: Optional. Delay, in seconds, between retries. Default 10


 *  `get.GetJsonCheck`

    Tests that gNMI Get of an xpath returns a schema-valid JSON response.

    Args:
     *  `xpath`: gNMI path to read.
     *  `model`: Python binding class to check the JSON reply against.
          The binding classes are in the `oc_config_validate.models` package.
     *  `retries`: Optional. Number of retries if the assertion fails.
     *  `retry_delay`: Optional. Delay, in seconds, between retries. Default 10.      


 *  `get.GetJsonCheckCompare`

    Checks for schema validity and compares a gNMI Get response.

    Args:
     *  `xpath`: gNMI path to read.
     *  `model`: Python binding class to check the JSON reply against.
          The binding classes are in the `oc_config_validate.models` package.
     *  `want_json`: Expected JSON-IETF value.
     *  `retries`: Optional. Number of retries if the assertion fails.
     *  `retry_delay`: Optional. Delay, in seconds, between retries. Default 10.

### Module set

 *  `set.SetUpdate`

    Sends gNMI Set Update of an xpath with a value.

    Args:
     *  `xpath`: gNMI path to update.
     *  `value`: Value to set; can be numeric, string or JSON-IETF.


 *  `set.SetDelete`

    Sends gNMI Set Delete of an xpath.

    Args:
     *  `xpath`: gNMI path to delete.


 *  `set.JsonCheckSetUpdate`

    Sends gNMI Set with a schema-checked JSON-IETF value.

    Args:
     *  `xpath`: gNMI path to read.
     *  `json_value`: JSON-IETF value to check and set.
     *  `model`: Python binding class to check the JSON reply against.
          The binding classes are in the `oc_config_validate.models` package.

### Module setget

 *  `setget.SetGetJsonCheck`

    Sends gNMI Set and later Get, schema-checking the JSON-IETF value.

     1.  The intended JSON-IETF configuration is checked for schema validity.
     1.  It is sent in a gNMI Set request.
     1.  The same path is used for a gNMI Get request.
     1.  The returned value is checked for schema validity

    Args:
     *  `xpath`: gNMI path to write and read.
     *  `json_value`: JSON-IETF value to check set and get.
     *  `model`: Python binding class to check the JSON reply against.
          The binding classes are in the `oc_config_validate.models` package.
     *  `retries`: Optional. Number of retries if the assertion fails.
     *  `retry_delay`: Optional. Delay, in seconds, between retries. Default 10.


 *  `setget.SetGetJsonCheckCompare`

    Does what setget.JsonCheck does, but also compares the JSON Get reply.

    1. The intended JSON-IETF configuration is checked for schema validity.
    1. It is sent in a gNMI Set request.
    1. The same path is used for a gNMI Get request.
    1. The returned value is checked for schema validity
    1. The returned value is compared with the sent value

    Args:
     *  `xpath`: gNMI path to write and read.
     *  `json_value`: JSON-IETF value to check set, get and compare.
     *  `model`: Python binding class to check the JSON reply against.
          The binding classes are in the `oc_config_validate.models` package.
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
     *  `model`: Python binding class to check the JSON reply against.
          The binding classes are in the `oc_config_validate.models` package.
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

    Args:
     * `prefix`: Destination prefix of the static route.
     * `index`: Index of the next hop for the prefix. Defaults to 0.


 *  `static_route.CheckRouteState`

     Checks the state on a static route.

    1. A gNMI Get message on the /state container.

    Args:
     * `prefix`: Destination prefix of the static route.
     * `next_hop`: IP of the next hop of the route.
     * `index`: Index of the next hop for the prefix. Defaults to 0.
     * `metric`: Optional numeric metric of the next hop for the prefix.
     * `description`: Optional text description of the route.


 *  `static_route.CheckRouteConfig`

     Checks the config on a static route.

    1. A gNMI Get message on the /config container.

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

    Args:
     * `interface`: Name of the physical interface.
     * `index`: Index of the subinterface, defaults to 0.
     * `dhcp`: True to enable DHCP, defaults to False.


 *  `subif_ip.CheckSubifDhcpState`

    Checks the DHCP state on a subinterface.

    1. A gNMI Get message on the /state container.

    Args:
     * `interface`: Name of the physical interface.
     * `index`: Index of the subinterface, defaults to 0.
     * `dhcp`: True to enable DHCP, defaults to False.


 *  `subif_ip.CheckSubifDhcpConfig`

    Checks the DHCP config on a subinterface.

    1. A gNMI Get message on the /config container.

    Args:
     * `interface`: Name of the physical interface.
     * `index`: Index of the subinterface, defaults to 0.
     * `dhcp`: True to enable DHCP, defaults to False.


 *  `subif_ip.AddSubifIp`

    Tests configuring an IP on a subinterface.

    1. A gNMI Set message is sent to configure the subinterface.
    1. A gNMI Get message on the /config container validates it.

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

    Args:
     * `interface`: Name of the physical interface.
     * `index`: Index of the subinterface, defaults to 0.
     * `address`: IPv4 address.
     * `prefix_length`: Prefix lenght of the IPv4 address.


 *  `subif_ip.CheckSubifIpState`

    Checks the state on an ip address configured on a subinterface.

    1. A gNMI Get message on the /state container.

    Args:
     * `interface`: Name of the physical interface.
     * `index`: Index of the subinterface, defaults to 0.
     * `address`: IPv4 address.
     * `prefix_length`: Prefix lenght of the IPv4 address.


 *  `subif_ip.CheckSubifIpConfig`

    Checks the configuration on an ip address on a subinterface.

    1. A gNMI Get message on the /config container.

    Args:
     * `interface`: Name of the physical interface.
     * `index`: Index of the subinterface, defaults to 0.
     * `address`: IPv4 address.
     * `prefix_length`: Prefix lenght of the IPv4 address.

### Module telemetry_once

Uses gNMI Subscribe messages, of type ONCE.

This testcase supports sending multiple xpaths in the gNMI Subscribe request.

Optionally, every Update messages can have its timestamp value checked
  against the local time when the Subscription message was sent. The absolute
  time drift is compared against a max value in secs.

Optionally, the number of expected Notifications (equals to the amount of)
  timestamps) is checked.

> If the arguments are not present in the test, the check is not performed.

Args:
  *  *notifications_count*: Number of expected Notification messages.
  *  *max_timestamp_drift_secs*: Maximum drift for the timestamp(s).


####  `telemetry_once.CountUpdatesCheckType`

In addition to the default checks of the module, this test checks that the
returned Updates and their values type, without checking any OC  model. This
is a rather basic test that relies on knowing exactly the expected replies to
the Subscription.

All Update values are expected to be of the same Type, as 'string_val',
  'int_val', etc.

Args:
 *  **xpaths**: List of gNMI paths to subscribe to. Paths can contain
     wildcard '*'.
 *  **updates_count**: Number of expected Update messages.
 *  **values_type**: Python type of the values of the Updates.


####  `telemetry_once.CheckLeafs`

In addition to the default checks of the module, this test checks that the
subscription to containers updates all leafs.

This test subscribes only xpaths of containers. It renders the
corresponding OC model and lists all paths to the downstream Leafs. The
Update paths received in the Subscription reply are checked against the
Leafs of the Model (Update paths must match an OC Model Leaf).

Optionally, use `check_missing_model_paths` to assert that all OC model paths
are present in the Updates. Usually, the Subscription replies might not
have all Leaf paths that the OC mode has.

> This check does NOT check the type of the values returned.

Args:
 *  **xpaths**: List of gNMI paths to subscribe to.
     Paths can contain wildcards only in keys '*'.
 *  **model**: Python binding class to check the reply against.
 *  *check_missing_model_paths*: If True, it asserts that all OC Model Leaf
     paths are in the received Updates. Defaults to False.

### Module telemetry_sample

Uses gNMI Subscribe messages, of type STREAM, mode SAMPLE.

This test subscribes to a single gNMI xpath, request sampled streaming
telemetry and keep accumulating responses up to a timeout.
After the timeout, it checks that all returned paths have the same number of
updates, and that the time interval between updates is the requested value
(with a maximum drift tolerance).

E.g:
Suppose the test is subscribing to an xpath `/<root>/<container>`, with 15 secs
interval for 65 secs. The Target replies with Updates for several paths
(`/<root>/<container>/<leaf*>`). After collecting subscription responses for
65 secs, the test checks that all paths have 5 updates, and for each path,
that the timestamp difference of updates is 15 secs.

 Update path | Time
 ----------- | ----
 `/<root>/<container>/<leaf1>` | Update[t0] >> Update[t15] >> Update[t30] >> ...
 `/<root>/<container>/<leaf2>` | Update[t0] >> Update[t18] >> Update[t25] >> ...

> These checks do not validate the values returned on the updates.
> Use telemetry_once.* for that.

Args:
 *  **sample_interval**: Seconds interval between Updates
 *  **sample_timeout**: Seconds to keep the Subscription and collect Updates.
 *  *max_timestamp_drift_secs*: Maximum drift for the timestamp(s) in the
    reply. Defaults to 1 sec.

####  `telemetry_sample.CountUpdates`

In addition to the default checks of the module, this test checks the
Subscription to the xpath produces Updates for a determined number of paths.

E.g:
Suppose the test is subscribing to an xpath `/<root>/<container>`, with 15 secs
interval for 65 secs, expecting 3 Update paths. After collecting subscription
responses for 65 secs, the test checks that there are Updates for 3 xpaths.

Update path | t0 | t15 | t30 | t45 | t60
----------- | -- | --- | --- | --- | ---
`/<root>/<container>/<leaf1>` | value | value | value | value | value
`/<root>/<container>/<leaf2>` |       | value | value |       |
`/<root>/<container>/<leaf3>` | value |       | value |       | value

Args:
 *  **xpath**: gNMI path to subscribe to. Path can contain wildcards '*'.
 * **update_paths_count**: Number of expected distinct Update paths.

####  `telemetry_sample.CheckLeafs`

In addition to the default checks of the module, this test checks that the
subscription to a container updates all leafs under it.

It renders the corresponding OC model and lists all paths to the downstream
Leafs. The Update paths received in the Subscription replies are checked
against the Leafs of the model (all Update paths must match an OC model Leaf).

 E.g:
 Suppose the test is subscribing to an xpath `/<root>/<container>/state`, with
 15 secs interval for 65 secs. After collecting subscription responses for 65
 secs, the test checks that all Update paths are valid in the given OC model,
 and that there are updates for all paths.

 Update path | t0 | t15 | t30 | t45 | t60
 ----------- | -- | --- | --- | --- | ---
 `/<root>/<container>/state/<leaf1>` | value | value | value | value | value
 `/<root>/<container>/state/<leaf2>` |       | value | value |       |
 `/<root>/<container>/state/<leaf3>` | value |       | value |       | value

> This check does NOT check the type of the values returned.

Args:
 *  **xpath**: gNMI path to subscribe to.
    Can contain wildcards only in keys '*'.
*  **model**: Python binding class to check the replies against.
