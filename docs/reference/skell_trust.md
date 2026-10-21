## skell trust

Choose which skill publishers you trust

### Synopsis

Manages the publishers whose signatures Skell accepts. The list lives in
trusted-keys.json in the Skell home, so a company can ship it with the rest of
its Skell configuration. With 'require-signed = true' in config.toml, only
skills signed by a trusted publisher can be installed or upgraded.

### Options

```
  -h, --help   help for trust
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.
* [skell trust add](skell_trust_add.md)	 - Trust a publisher's public key
* [skell trust list](skell_trust_list.md)	 - Show trusted publishers
* [skell trust remove](skell_trust_remove.md)	 - Stop trusting a publisher

