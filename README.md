# Fluent Bit X Sumo Logic

Output plugin for `Fluent Bit` which can send logs to `Sumo Logic`!

## Usage

Build the code in this repository as a `c-shared` object, using the following command:

```bash
go build -buildmode=c-shared -o out_sumologic.so main.go
```

Run `Fluent Bit` with this new plugin, using the following command:

```bash
fluent-bit -e out_sumologic.so -i tail -o sumologic
```

Refer to the [Fluent Bit Docs](https://docs.fluentbit.io/manual/development/golang-output-plugins) and this [Dockerfile](./Dockerfile) for more hints.


## Configuration

This plugin supports the following configuration properties:

| Property           | Required |     Default Value    | Tag Substitution |  Description |
|:-------------------|:--------:|:--------------------:|:----------------:|:-------------|
| `Collector_Url`    |    ✅     |         ❌           |        ❌        | `Sumo Logic` Hosted Collector Endpoint for Logs. For more details refer to - [Sumo Logic Docs](https://help.sumologic.com/docs/send-data/hosted-collectors/configure-hosted-collector/) |
| `Match`            |    ✅     |         ❌           |        ❌        | Used to route data between the input and output plugins in `Fluent Bit`. For more details refer to - [Fluent Bit Docs](https://docs.fluentbit.io/manual/concepts/data-pipeline/router) |
| `Source_Name`      |    ❌     |    Unconfigured     |        ✅        | Specify an override for the Source Name. Defaults to the Collector Source Name configured in `Sumo Logic` |
| `Source_Host`      |    ❌     |   `os.Hostname()`   |        ✅        | Specify an override for the Source Host. Defaults to the hostname returned by the kernel |
| `Source_Category`  |    ❌     | `sumologic_default` |        ✅        | Specify the Source Category to send logs to in `Sumo Logic` |
| `Tag_Delimiter`    |    ❌     |        `.`          |        ❌        | Used to split the tag sent by the input plugin, helps with the `Tag Substitution` |
| `Level`            |    ❌     |       `info`        |        ❌        | Specify the log level for the plugin. Defaults to `info` |
| `Log_Key`          |    ❌     |       `log`         |        ❌        | Specify the field to be extracted from the json record, sends the entire record to `Sumo Logic` otherwise. |
| `Max_Retries`      |    ❌     |       `10`          |        ❌        | Max retries in case of failures when attempting to communicate with `Sumo Logic` |


## Tag Substitution

The plugin supports dynamically populating the supported configuration options with values from the tag sent by the input plugin.

Example configuration

```ini
[INPUT]
    Name                tail
    Path                /var/log/containers/*.log
    Parser              docker
    Tag                 kube.<namespace_name>.<pod_name>.<container_name>
    Tag_Regex           (?<pod_name>[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*)_(?<namespace_name>[^_]+)_(?<container_name>.+)-

[OUTPUT]
    Name                sumologic
    Match               kube.*
    Collector_Url       url
    Tag_Delimiter       .
    Source_Host         $TAG[1]_$TAG[2]
    Level               info
```

In the above configuration, the tag is populated with the information about the `Kubernetes Namespace`, `Pod Name`, and `Container Name` by the input plugin.

This tag is then split by the `sumologic` plugin with the help from the `Tag_Delimiter` property to create the `TagSlice` and replaces all the occurrences of `$TAG[i]` (where `i` is the index of the element in `TagSlice`)

So in this case `Source_Host` would be evalutated as `<namespace_name>_<pod_name>`, which is very helpful in environments like Kubernetes.