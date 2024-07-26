## Using a custom binding in oc_config_validate

Sometimes the public OpenConfig models are not sufficient for all test cases.
 A few scenarios could be:

1.  A vendor creating a vendor specific model, which may include
 [vendor augments](https://github.com/openconfig/public/blob/master/doc/vendor_counter_guide.md).

1.  A vendor contributing on a public model change while developing software
 that supports the changes.

1.  A user that would like to test a vendor specific model implementation, or
 would like to test with different sets of public model [releases](https://github.com/openconfig/public/releases).

### Understanding the use case

This example will use this vendor specific model: https://github.com/aristanetworks/yang/blob/master/WIFI-16.1.0/release/openconfig/models/arista-wifi-augments.yang

This is a model for Arista Access Points.  It contains specific implementations
only relevant to this vendor. This model does a couple of things:

1.  Creates a top-level `arista-aps` grouping.
1.  *Augments* the OpenConfig `access-points` model.  

One leaf specifically, `tx-beacon` is interesting because it is a state-only
 leaf for an SSID, the most commonly configured setting on an AP.

This leaf was added as the public model implementation was still in progress,
thus this leaf did not exist on the public models.

With that in mind, given the following testcase:
```
- !TestCase
  name: Check base model with baseline (init) configuration.
  class_name: get.GetJsonCheck
  args:
    xpath: "/access-points/access-point[hostname=ap.example.com]/"
    model: "wifi.openconfig_access_points"
```

The following error is raised.

```
AssertionError: False is not true : Get response JSON does not match model: JSON object contained a key that did not exist (tx-beacon)
```

This makes sense.  The init configuration has [SSIDs configured](https://github.com/google/gnxi/blob/master/oc_config_validate/init_configs/ap.json),
 and when verifying the `get` to `/access-points/access-point[hostname=ap.example.com]/`
 against the public models binding, it fails.

In addition, we are not able to test the `spectratalk-pmk` leaf because the
 public models binding is not aware of the `arista-aps` grouping.  Ideally we
 want to have a `SetConfigCheckState` testcase that configures and checks this
 leaf.  

#### Generating the custom binding

We need a custom binding for our tests above to:

1. Validate against the custom binding
1. Test the vendor-specific config leafs, eg. `update` this path `/arista-aps/arista-ap[mac=aa:bb:cc:dd:ee:ff]/system/config/spectratalk-pmk`

An example of generating the binding is on [update_models.sh](https://github.com/google/gnxi/blob/master/oc_config_validate/oc_config_validate/update_models.sh#L43)

Notice how in addition to [wifi.py](https://github.com/google/gnxi/blob/master/oc_config_validate/oc_config_validate/models/wifi.py)
 we also generate [wifi_arista.py](https://github.com/google/gnxi/blob/master/oc_config_validate/oc_config_validate/models/wifi_arista.py)
 which incorporates the vendor specific model.

#### Using a vendor specific model

We can now use our custom binding in our models folder.

Our testcase should reference `wifi_arista.openconfig_access_points`:

```
- !TestCase
  name: Check base model with baseline (init) configuration.
  class_name: get.GetJsonCheck
  args:
    xpath: "/access-points/access-point[hostname=ap.example.com]/"
    model: "wifi_arista.openconfig_access_points"
```

And our test is now passing.

We can also have a testcase for the `spectratalk-pmk` leaf. This leaf is
is not an augment, but it is specific to arista-wifi-augments model.

```
- !TestCase
  name: "Set PMK value to test hex string"
  class_name: config_state.SetConfigCheckState
  args:
    xpath: "/arista-aps/arista-ap[mac=aa:bb:cc:dd:ee:ff]/system/"
    json_value: {
      "arista-wifi-augments:mac": "aa:bb:cc:dd:ee:ff",
      "arista-wifi-augments:spectratalk-pmk": "90cb93e70f52cc17fb9a3bc68940aa95"
    }
    model: "wifi_arista.arista_wifi_augments"
```

Note how the model is being referenced as `wifi_arista.arista_wifi_augments`,
the custom binding name followed by the model name itself.
