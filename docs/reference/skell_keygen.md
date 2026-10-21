## skell keygen

Create a signing key for publishing skills

### Synopsis

Creates an Ed25519 key pair for a publisher (your name, team or company):
<publisher>.key is private — keep it secret and never commit it — and
<publisher>.pub is the public half you hand to people who want to trust your
skills ('skell trust add').

Keys are written to --out, by default the 'keys' folder inside the Skell home.
An existing key is never overwritten.

```
skell keygen <publisher> [flags]
```

### Examples

```
  skell keygen acme
  skell keygen acme --out ~/secure/keys
```

### Options

```
  -h, --help         help for keygen
      --out string   Folder to write the key files to
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

