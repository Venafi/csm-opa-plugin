# CyberArk Code Sign Manager Plugin for Open Policy Agent

This OPA plugin for Code Sign Manager (previously CodeSign Protect) provides bundle signing and verification functionality.

Starting at `cmd/opa/main.go`, we hook into OPA's RootCommand and inject a PersistentPreRun 
hook for certain OPA commands. We call `bundle.RegisterSigner` and `bundle.RegisterVerifier` 
for our custom implementations of the `bundle.Signer` and `bundle.Verifier` interfaces, respectively.

### Configuration

This plugin relies on environment variables, and therefore must be set prior to running opa with the plugin.  Review the [vSign](https://github.com/Venafi/vsign) SDK for detailed information on creating the necessary Venafi API oauth token.

#### Create Environment Variables

These are the minimum variables required

```sh
VSIGN_URL="https://tpp.example.com"
VSIGN_TOKEN="xxxxxxxxxx"
VSIGN_JWT="xxxxxxxxxxx"
```

For authentication only use either `VSIGN_TOKEN` or `VSIGN_JWT`, since the JWT will be exchanged for an access token.

*Currently only Certificate environments are supported*

### Signing and Running

```sh
./bin/opa build --bundle ./policy --output ./policy/bundle.tar.gz --signing-key vsign\\rsa2048-cert --signing-plugin csm-opa-plugin
./bin/opa sign --bundle --signing-key vsign\\rsa2048-cert --signing-plugin csm-opa-plugin ./policy
```

```sh
./bin/opa run --bundle --verification-key vsign\\rsa2048-cert --verification-key-id vsign\\rsa2048-cert --exclude-files-verify data.json --exclude-files-verify policy/awesome.rego --exclude-files-verify .manifest --exclude-files-verify .signatures.json ./policy/bundle.tar.gz
```
