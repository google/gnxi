## Vendor Testing Procedure

### Summary

This document describes the recommended guidelines that network vendors can follow to run **oc\_config\_validate** continuously and effectively share results with customers.  These guidelines are intended to foster collaboration and transparency between the two parties.

---

The following is a list of recommendations – which are explained in detail individually.

1. Use and augment customer test plans
2. Keep tests current
3. Use current models
4. Integrate oc\_config\_validate into QA pipeline as a continuous integration
5. Share results, test and init\_config files

### Use and augment per-customer configurations

Ideally, customers would have initial configurations (init\_configs) and tests files (test.yaml) that are representative of the features in use and the expectations of how devices should perform, per hardware platform – called the **test plans**.

Customer-provided test plans should serve as a _baseline_, and vendors should be able to augment these with their own tests.

If a customer does not have a test plan, vendors should be able to create test plans that cater to the known customer needs, eg. hardware platforms in use, features used, software-releases, etc.

### Keep tests current

It is expected that with different software releases or hardware, the test plans may differ and thus it is important to stay current.  Vendors should work with customers to keep test plans up-to-date, and/or update their own test plans as new features are added and supported.  

### Use current models

Since OpenConfig models change over time – it is important to keep up-to-date with the latest models.  As OC models can change and include non-backwards compatible changes, it is important that devices adhere to schema.

The OpenConfig models release versioning and release timing is described [here](https://github.com/openconfig/public/blob/master/doc/releases.md).  The oc\_config\_validate PyPi package uses the latest models bindings at the time of release.  Running `oc\_config\_validate -models` prints the OpenConfig models versions used.  Optionally,  custom model bindings can be generated when running oc\_config\_generate from source, to use specific OC models versions  – read more about updating model bindings or using custom models [here](https://github.com/google/gnxi/blob/master/oc_config_validate/docs/oc_models.md#openconfig-oc-models).


### Integrate oc\_config\_validate into QA pipeline

oc\_config\_validate is most valuable when used as part of a release pipeline, ideally running in continuous fashion rather than as a one-shot.  There are several benefits to running in this fashion, including:

1. Assurance of consistency
2. Reference historical timing for tests
3. Discovery of flaky tests or conditions
4. Ability to have long-running tests
5. Greater overall degree of confidence in results

### Share results, test and init\_config files

Having the same point of reference, and sharing the results helps to create better collaboration.  

1. Disclose any issues found during testing
2. Share a raw results.json that is representative of the state of the tests
    1. If using multiple tests files (\*.yaml), share each results.json file for the given test .yaml file
3. Provide all initial configuration files (init\_configs/\*.json) used as part of the tests
4. Share all test files (\*.yaml)
    2. This should include all tests run.  If using customer-provided test files, inform the customer of any alterations to the tests and share the updated files back.
    3. The specific logic of the tests is important. For instance, if a test requires a long timeout or retries, this should be clear from the test files.

A possible approach for vendors is to create a small directory structure per customer:

```

/customer1

  ./tests/\*.yaml

  ./init\_configs/\*.json

  ./results/\*.json

/customer2

  ./tests/\*.yaml

  ./init\_configs/\*.json

  ./results/\*.json

 ```

And calling oc\_config\_validate with arguments pointing to the per-customer path.

Moreover, these directories can be tracked in a private VCS repo or in a shared file system, so customers can easily consume them.
