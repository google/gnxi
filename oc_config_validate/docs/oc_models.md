## OpenConfig (OC) models

### Building and updating the OC models bindings

**oc_config_validate** includes the Python bindings of most OpenConfig models, in the package `oc_config_validate.models`.

This package is built and updated using the **oc_config_validate/update_models.sh** script, which does:

1. Delete the current Python package.
1. Get a copy of the [OpenConfig public repository](https://github.com/openconfig/public.git).
1. For every folder in the repo's `/public/models` path, a Python module is created with all the YANG models contained in the folder.
1. The revisions of the binded YANG models is collected in `oc_config_validate/oc_config_validate/models/versions`.

> The structure of the Python package `oc_config_validate.models` is the same as the folder structure of https://github.com/openconfig/public/tree/master/release/models.

### Using the OC models in testclasses

The model bindings are used in tests that validate the gNMI JSON payloads against the expected OpenConfig model.

 *  As test arguments for basic testclasses, like `config_state.SetConfigCheckState`:

    The `model` class argument is expected to be in the form *module.class* referenced to the `oc_config_validate.models` package. There is no need to examine the Python code if one remembers that the same folder structure of the OC models public repo is mirrored.

    I.e.: to use the *openconfig_interfaces* OC model, found in https://github.com/openconfig/public/blob/master/release/models/interfaces/openconfig-interfaces.yang, the testclass argument is `model: interfaces.openconfig_interfaces`. This will, in the end, resolve to the Python class `oc_config_validate.models.interfaces.openconfig_interfaces`.

    Of course, the tests use rather a section of the whole OC model, referenced by the `xpath` parameter pointing to a container element. During the tests, **oc_config_validate** will infer the Python class corresponding to the container element from the `xpath` argument, in the given `model` model.

    I.e.: to check the model validity of the container pointed by `xpath: /system/clock/config` in the given model `model: system.openconfig_system`, **oc_config_validate** will infer and use the Python class `oc_config_validate.models.interfaces.openconfig_system.yc_config_openconfig_system__system_clock_config`.

 *  As a Python import for advanced testclasses, like `subif_ip.CheckSubifIpConfig`:

    Simply import the OC model from the Python *module*, which corresponds to the folder in https://github.com/openconfig/public/blob/master/release/models where the model is.

    E.g.: `from oc_config_validate.models.interfaces import openconfig_if_ip`.

    You will then be able to create Python objects that correspond to the OC model containers.

### Using custom OC models

Technically, any YANG OC model that is binded to a Python class in the `oc_config_validate.models` package can be used in testclasses. However, we suggest 2 approaches:

 *  Create a `custom` Python module:

    1.  Put a few custom YANG OC models (either previous revisions, or other models altogether) in a folder named `custom` or something similar.
    1.  Use `pyang` to create a single Python file `oc_config_validate/models/custom` package that contains the custom models. See **oc_config_validate/update_models.sh**.
    1.  In the testclasses, reference `oc_config_validate.models.custom` module or `custom.openconfig_<model>` model class.

 *  Replace the `oc_config_validate.models` Python package:

    1.  Put all YANG OC models (either previous revisions, or other models altogether) is a folder structure similar to https://github.com/openconfig/public/blob/master/release/models.
    1.  Use **oc_config_validate/update_models.sh** to create all the code bindings from your YANG model folder, instead of fetching them from upstream.
    1.  In the testclasses, reference `oc_config_validate.models.<module>` module or `<module>.openconfig_<model>` model class.
