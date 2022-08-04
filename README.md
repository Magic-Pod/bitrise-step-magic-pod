# magic-pod

This step enables E2E testing powered by [MagicPod](https://magicpod.com)


## How to use this Step

Can be run directly with the [bitrise CLI](https://github.com/bitrise-io/bitrise),
just `git clone` this repository, `cd` into it's folder in your Terminal/Command Line
and call `bitrise run test`.

*Check the `bitrise.yml` file for required inputs which have to be
added to your `.bitrise.secrets.yml` file!*

Requirements:

- You need to sign up to `https://app.magicpod.com` and create the following.
  - Project
  - Test cases
  - Test settings (which defines how batch runs should be executed)
- You also need to confirm your API token in `https://app.magicpod.com/accounts/api-token/`

Step by step:

1. Open up your Terminal / Command Line
2. `git clone` the repository
3. `cd` into the directory of the step (the one you just `git clone`d)
5. Create a `.bitrise.secrets.yml` file in the same directory of `bitrise.yml`
   (the `.bitrise.secrets.yml` is a git ignored file, you can store your secrets in it)
6. Check the `bitrise.yml` file for any secret you should set in `.bitrise.secrets.yml`
  * Best practice is to mark these options with something like `# define these in your .bitrise.secrets.yml`, in the `app:envs` section.
7. Once you have all the required secret parameters in your `.bitrise.secrets.yml` you can just run this step with the [bitrise CLI](https://github.com/bitrise-io/bitrise): `bitrise run test`

An example `.bitrise.secrets.yml` file:

```
envs:
- MAGIC_POD_API_TOKEN: "<YOUR_TOKEN>"
- ORGANIZATION_NAME: "<YOUR_ORGANIZATION_NAME>"
- PROJECT_NAME: "<YOUR_PROJECT_NAME>"
- TEST_SETTINGS_NUMBER: "<YOUR_TEST_SETTINGS_NUMBER>"
- TEST_SETTINGS: ""
- APP_PATH: "<PATH_TO_YOUR_APP>"
- WAIT_FOR_RESULT: "true"
- WAIT_LIMIT: "0"
- DELETE_APP_AFTER_TEST: "Not delete"
```
