## skell sign

Sign a skill (or every skill in a folder) so others can verify who published it

### Synopsis

Writes SKILL.sig into a skill folder. The signature covers every file in the
skill, so any later change — by anyone — makes 'skell verify' report it as
tampered. Re-sign after every edit, ideally as part of your release.

<path> is a skill folder, or a folder containing several skills (a source
repository), in which case all of them are signed.

```
skell sign <path> [flags]
```

### Examples

```
  skell sign ./release-notes --key ~/.skell/keys/acme.key --publisher acme
  skell sign ~/src/team-skills --key ~/.skell/keys/acme.key --publisher acme
```

### Options

```
  -h, --help               help for sign
      --json               Output results as JSON
      --key string         Private key file from 'skell keygen'
      --publisher string   Publisher name recorded in the signature
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

