## skell bundle

Named sets of skills (a role or team starter pack) you install in one step

### Synopsis

A bundle is a named list of skills stored in the project's skell.toml:

  [bundles.frontend]
  description = "Everything a UI developer needs"
  skills = ["react", "css-review", "a11y-check"]

Install one with 'skell bundle install frontend' (or 'skell install @frontend').
Bundles travel with 'skell apply <baseline.toml>', so a team can publish its
starter packs once and every project picks them up. Dependencies declared by
the skills (metadata.requires) are installed along with them.

### Options

```
  -h, --help   help for bundle
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.
* [skell bundle create](skell_bundle_create.md)	 - Create (or replace) a bundle
* [skell bundle delete](skell_bundle_delete.md)	 - Delete a bundle (its skills stay installed)
* [skell bundle install](skell_bundle_install.md)	 - Install every skill in one or more bundles
* [skell bundle list](skell_bundle_list.md)	 - Show the bundles defined in this project

