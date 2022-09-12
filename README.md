# Grafana Backend Datasource Plugin for Db2 for z/OS

This plugin allows the visualisation of data contained in Db2 for z/OS tables through Grafana.

## Note

Backend datasource plugins run a tertiary program inside of your Grafana instance. Therefore, backend datasource plugins must be signed to make sure they have not been tampered with. Signing plugins is not yet available to individuals such as myself, only Grafana Labs or Enterprise partners can produce signed plugins. You can allow Grafana to load unsigned plugins, but I strongly advise against this unless you inspect the code in this repository first, and then build the plugin yourself. In order for Grafana to load this plugin, add the following line to your Grafana /conf/config.ini;

```
allow_loading_unsigned_plugins = "jcnnrts-db-2-datasource-v-8"
```

If you need help building, or absolutely want a pre-built /dist folder, send me a message.

## Building

The build process is only tested on Windows, for Windows. The go_ibm_db package can be had for Linux and Darwin as well, but the Grafana build process doesn't play nice with cross-compiling. Change Magefile.go to attempt to build for another platform.

### Tools needed
- go
- mage
- yarn

### Go dependencies needed for building:

The Grafana plugin SDK:
```
go get -u github.com/grafana/grafana-plugin-sdk-go
```

Db2 clidriver and its Go wrapper (Windows):

Originally by the IBMDB account on github, forked by me because the pooling code was sketchy at best.
```
go get -d github.com/jcnnrts/go_ibm_db
cd %GOPATH%\src\github.com\jcnnrts\go_ibm_db\installer
go run setup.go
```

### Build

You want to do this in the /plugins folder of your Grafana installation unless you altered your custom.ini file to load plugins from another location.

```
git clone https://github.com/jcnnrts/db-2-datasource-v-8
cd db-2-datasource-v-8
yarn install
yarn build
mage -v
```