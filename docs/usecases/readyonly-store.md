# Use case: Readonly Store

## Summary

Allow a password store or a set of sub-stores to be configured in readonly mode in team sharing scenario.

## Background

In a team sharing scenario that we share password store among team members, usually we want each member to be able to pull and push so that anyone can share secret data with others. However, typically in a large size team, we may want the sharing to be more restricted where only priviledged users are allowed to push secret data to the remote store while others can only pull from the remote store.

The current gopass behavior is that it will auto sync (pull and push) betweeen local store and remote store any time when there is a change at local. This means anyone can push their personal data to the centrally controlled remote store which will be polluted with arbitrary data unexpectedly.

To workaround this, we can configure "Collaborators & teams" on GitHub side to grant read only permission to those who do not necessarily need push, but the gopass on their machines will still keep auto pushing and prompt with errors which is annoying.

## Proposal

Ultimately it turns out that this scenario requires a feature such as a store in readonly mode, where people can configure their local store or a set of sub-stores in readonly mode, to disable the writes and the autosync-on-writes to the store, but they can still pull to sync the latest changes from the remote store. This is not a one-stop solution for the RBAC model of the team sharing store, because we still need GitHub to setup the store access at server-side, but it will provide better usage experience from gopass client side.

Configuration examples:
```console
# To print the config
$ gopass config readonly
readonly: false
# To setup the config
$ gopass config readonly true
readonly: true
# To apply the config to a sub-store
$ gopass config readly true --store team-sharable
readonly: true
```

## References

* https://github.com/gopasspw/gopass/issues/1878
