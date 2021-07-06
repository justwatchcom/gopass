# Features

This document provides a broad overview over the features and use-cases
gopass supports.

## Standard Features

Note: Running `gopass` without any arguments opens up an interactive mode where
all commands explained below are available without the need to prefix them with
`gopass`. Also this mode offers tab completion without the need to configure
the shell.

### Data Organization

Before you start using gopass, you should know a little bit about how it stores your data.
It's actually really simple! Each password (or secret) will live in its own file.
And you can stick related passwords (or secrets) together in a directory.
So, for example, if you had 3 laptops and wanted to store the root passwords for all 3, then your file system might look something like the following:

```
.password-store
└── laptops
    ├── dell.gpg
    ├── hp.gpg
    └── macbook.gpg
```

With this file system, if you typed the `gopass ls` command, it would report the following:

```
gopass
└── laptops
    ├── dell
    ├── hp
    └── macbook
```

In this example, the key for the MacBook is `laptops/macbook`.

gopass does not impose any specific layout for your data. Any key can contain any kind of data. Please note that sensitive data **should not** be put into the name of a secret.

If you plan to use the password store for website credentials or plan to use [browserpass](https://github.com/dannyvankooten/browserpass), you should follow the following pattern for storing passwords:

```
example1.com/username
example2.com/john@doe.com
```

### Initializing a Password Store

After installing gopass, the first thing you should do is initialize a password store.
(If you are migrating to gopass from pass and already have a password store, you can skip this step.)

Note that this document uses the term *password store* to refer to a directory that is managed by gopass.
This is entirely different from any OS-level credential store, your GPG key ring, or your SSH keys.

To initialize a password store, just do:

```bash
gopass init
```

This will prompt you for which GPG key you want to associate the store with.
Then it will create a `.password-store` directory in your home directory.

If you don't want gopass to use this default directory, you can instead initialize a password store with:

```bash
gopass init --path /custom/path/to/password/store
```

If you don't want gopass to prompt you for the GPG key to use, you can specify it inline. For example, this might be useful if you have a huge number of GPG keys on the system or if you are initializing a password store from a script. You can do this in three different ways:

```bash
gopass init gopher@golang.org # By specifying the email address associated with the GPG key
gopass init A3683834 # By specifying the 8 character ID found by typing "gpg --list-keys" and looking at the "pub" line
gopass init 1E52C1335AC1F4F4FE02F62AB5B44266A3683834 # By specifying the GPG key fingerprint found by typing "gpg --fingerprint" and removing all of the spaces
```

### Cloning an Existing Password Store

If you already have an existing password store that exists in a Git repository, then use `gopass` to clone it:

```bash
gopass clone git@example.com/pass.git
```

This runs `git clone` in the background. If you don't want gopass to use the default directory of "$HOME/.password-store", then you can specify an additional parameter:

```bash
gopass clone git@example.com/pass-work.git work # This will initialize the password store in the "$HOME/.password-store-work" directory
```

Please note that all cloned repositories must already have been initialized with gopass. (See the previous section for more details.)

Note too that unless you are already a recipient of the cloned repository, you must add the destination's public GPG key as a recipient to the existing store.

### Adding Secrets

Let's say you want to create an account.

| Website    | User   |
| ---------- | ------ |
| golang.org | gopher |

#### Type in a new secret

```bash
$ gopass insert golang.org/gopher
Enter secret for golang.org/gopher:       # hidden on Linux / MacOS
Retype secret for golang.org/gopher:      # hidden on Linux / MacOS
gopass: Encrypting golang.org/gopher for these recipients:
 - 0xB5B44266A3683834 - Gopher <gopher@golang.org>

Do you want to continue? [yn]: y
```

#### Generate a new secret

```bash
$ gopass generate golang.org/gopher
How long should the secret be? [20]:
gopass: Encrypting golang.org/gopher for these recipients:
 - 0xB5B44266A3683834 - Gopher <gopher@golang.org>

Do you want to continue? [yn]: y
The generated secret for golang.org/gopher is:
Eech4ahRoy2oowi0ohl
```

```bash
$ gopass generate golang.org/gopher 16    # length as parameter
gopass: Encrypting golang.org/gopher for these recipients:
 - 0xB5B44266A3683834 - Gopher <gopher@golang.org>

Do you want to continue? [yn]: y
The generated password for golang.org/gopher is:
Eech4ahRoy2oowi0ohl
```

The `generate` command will ask for any missing arguments, like the name of the secret or the length. By default the password is copied to clipboard. If you don't want the password to be copied, but displayed instead, use the `-p` flag to print it.

By default the password is copied to clipboard, but you can disable this using the `AutoClip` option, which, when set to`false`, will neither display, nor print the password. This is overridden by the `-p` or `-c` flags.

### Edit a secret

```bash
$ gopass edit golang.org/gopher
```

The `edit` command uses the `$EDITOR` environment variable to start your preferred editor where you can easily edit multi-line content. `vim` will be the default if `$EDITOR` is not set.

### Adding OTP Secrets

Warning: OTP support is deprecated.

*Note: Depending on your security needs, it may not behoove you to store your OTP secrets alongside your passwords! Look into [Multiple Stores](https://github.com/gopasspw/gopass/blob/master/docs/features.md#multiple-stores) if you need things to be separate!*

Typically sites will display a QR code containing a URL that starts with `oauth://`. This string contains information about generating your OTPs and can be directly added to your password file. For example:

```
gopass show golang.org/gopher
secret1234
otpauth://totp/golang.org:gopher?secret=ABC123
```

Alternatively, you can use YAML (currently totp only):

```
gopass show golang.org/gopher
secret1234
---
totp: ABC123
```

*Note: any values for `totp:` need to be base32 (32, not 64) encoded. Often sites will display the raw secret alongside the QR*

Some sites will not directly show you the URL contained in the QR code. If this is the case, you can use something like [zbar](http://zbar.sourceforge.net/) to extract the URL.

Both TOTP and HOTP are supported. However, to generate HOTP tokens, the counter in the stored URL must be manually incremented (e.g. via `gopass edit myhotpsecret`). 

### Listing existing secrets

You can list all entries of the store:

```bash
$ gopass ls
gopass
├── golang.org
│   └── gopher
└── emails
    ├── user@example.com
    └── user@justwatch.com
```

If your terminal supports colors the output will use ANSI color codes to highlight directories and mounted sub stores. Mounted sub stores include the mount point and source directory. See below for more details on mounts and sub stores.

### Show a secret

```bash
$ gopass show golang.org/gopher

Eech4ahRoy2oowi0ohl
```

The default action of `gopass` is show, so the previous command is exactly the same as typing `gopass golang.org/gopher`. It also accepts the `-c` flag to copy the content of the secret directly to the clipboard.

In order to display only the password, the flag `-o` can be used. One can also copy the password to the clipboard while still showing the entry using the flag `-C`.

WARNING: The short form `gopass <secret>` is deprecated. Use `gopass show <secret>`.

Since it may be dangerous to always display the password, the `safecontent` setting may be set to `true` to allow one to display only the rest of the password entries by default but hiding the password. In order to display the whole entry, with the password in clear, the `-u`/`--unsafe` flag must then be used.
The password can still be shown using the `-o` flag.

WARNING: The `safecontent` setting is not perfect and *might* be removed in the future.

#### Copy a secret to the clipboard

```bash
$ gopass show -c golang.org/gopher

Copied golang.org/gopher to clipboard. Will clear in 45 seconds.
```

### Removing a secret

```bash
$ gopass rm golang.org/gopher
```

`rm` will remove a secret from the store. Use `-r` to delete a whole folder. Please note that you **can not** remove a folder containing a mounted sub store. You have to unmount any mounted sub stores first.

### Moving a secret

```bash
$ gopass mv emails/example.com emails/user@example.com
```

*Moving also works across different sub-stores.*

### Copying a secret

```bash
$ gopass cp emails/example.com emails/user@example.com
```

*Copying also works across different sub-stores.*

## Advanced Features

### Auto-Pager

Like other popular open-source projects, gopass automatically pipe the output to `$PAGER` if it's longer than one terminal page. You can disable this behavior by unsetting `$PAGER` or `gopass config nopager true`.

### Sync

Gopass offers as simple and intuitive way to sync one or many stores with their
remotes. This will perform and git pull, push and import or export any missing
GPG keys.

```bash
$ gopass sync
```

### Desktop Notifications

Certain long running operations, like `gopass sync` or `copy to clipboard` will
try to show desktop notifications [Linux only].

### git auto-push and sync

gopass always pushes changes to your default git remote server (origin).

If you want to pull changes from git, you need to run the sync command:

```bash
$ gopass sync 
```

You can selectively pull changes into named stores:

```bash
$ gopass sync --store foo 
```

For details see: [`sync` command](commands/sync.md)

### Check Passwords for Common Flaws

gopass can check your passwords for common flaws, like being too short or coming from a dictionary.

```bash
$ gopass audit
Detected weak secret for 'golang.org/gopher': Password is too short
```

### Check Passwords against leaked passwords

gopass can assist you in checking your passwords against those included in recent data breaches.
You can either check against the HIBPv2 API (recommended) or download the dumps (v1 or v2) and
perform the check fully offline.

#### Using the API

```bash
gopass audit hibp --api
```

#### Using the Dumps

First go to [haveibeenpwned.com/Passwords](https://haveibeenpwned.com/Passwords) and download the dumps. Then unpack the 7-zip archives somewhere. Note that full path to those files and provide it to gopass `--dumps` flag.

```bash
$ gopass audit hibp --dumps /tmp/pwned-passwords-1.0.txt
```

### Support for Binary Content

WARNING: Binary support is undergoing changes. Expect changes to these commands.

gopass provides secure and easy support for working with binary files through the `cat`, `fscopy`, `fsmove` and `sum` family of sub-commands. One can copy or move secret from or to the store. gopass will attempt to securely overwrite and remove any secret moved to the store.

```bash
# copy file "/some/file.jpg" to "some/secret" in the store
$ gopass fscopy /some/file.jpg some/secret
# move file "/home/user/private.key" to "my/private.key", removing the file on disk
# after the file has been encoded, stored and verified to be intact (SHA256)
$ gopass fsmove /home/user/private.key my/private.key
# Calculate the checksum of some asset
$ gopass sha256 my/private.key
```

### Multiple Stores

gopass supports multi-stores that can be mounted over each other like file systems on Linux/UNIX systems. Mounting new stores can be done through gopass:

```bash
# Mount a new store
$ gopass mounts add test /tmp/password-store-test
# Show mounted stores
$ gopass mounts
# Unmount a store
$ gopass mounts remove test
```

You can initialize a new store using `gopass init --store mount-point --path /path/to/store`.

Where possible sub stores are supported transparently through the path to the secret. When specifying the name of a secret it's matched against any mounted sub stores and the given action is executed on this store.

Commands that don't accept an secret name, e.g. `gopass recipients add` or `gopass init` usually accept a `--store` parameter. Please check the help output of each command for more information, e.g. `gopass help init` or `gopass recipients help add`.


Commands that support the `--store` flag:

| **Command**                | **Example**                                   | **Description** |
| -------------------------- | --------------------------------------------- | --------------- |
| `gopass git push`          | `gopass git push --store=foo origin master`   | Push all changes in the sub store *foo* to master |
| `gopass git pull`          | `gopass git pull --store=foo origin master`   | Pull all changes in the sub store *foo* from master |
| `gopass git init`          | `gopass git init --store=foo`                 | Initialize git in the sub store *foo* |
| `gopass init`              | `gopass init --store=foo`                     | Initialize and mount the new sub store *foo* |
| `gopass recipients add`    | `gopass recipients add --store=foo GPGxID`    | Add the new recipient *GPGxID* to the store *foo* |
| `gopass recipients remove` | `gopass recipients remove --store=foo GPGxID` | Remove the existing recipients *GPGxID* from the store *foo* |

### Directly edit structured secrets aka. YAML support

Warning: YAML support is deprecated.

gopass supports directly editing structured secrets (simple key-value maps or YAML).

```bash
$ gopass generate -n foo/bar 12
The generated password for foo/bar is:
7fXGKeaZgzty
$ gopass insert foo/bar baz
Enter password for foo/bar/baz:
Retype password for foo/bar/baz:
$ gopass foo/bar baz
zab
$ gopass foo/bar
7fXGKeaZgzty
baz: zab
```

Please note that gopass will try to leave your secret as is whenever possible,
but as soon as you mutate the YAML content through gopass, i.e. `gopass insert secret key`,
it will employ an YAML marshaler that may alter the order and escaping of your
entries.

### Edit the Config

gopass allows editing the config from the command-line. This is similar to how git handles config changes through the command-line. Any change will be written to the configured gopass config file.

```bash
$ gopass config
askformore: false
autoclip: true
autoimport: false
cliptimeout: 10
noconfirm: false
path: /home/user/.password-store

$ gopass config cliptimeout 60
$ gopass config cliptimeout
```

### Managing Recipients

You can list, add and remove recipients from the command-line.

```bash
$ gopass recipients
gopass
└── 0xB5B44266A3683834 - Gopher <gopher@golang.org>

$ gopass recipients add 1ABB2C1A

$ gopass recipients
gopass
├── 0xB1C7DF661ABB2C1A - Someone <someone@example.com>
└── 0xB5B44266A3683834 - Gopher <gopher@golang.org>

$ gopass recipients remove 0xB5B44266A3683834

$ gopass recipients
gopass
└── 0xB1C7DF661ABB2C1A - Someone <someone@example.com>
```

Running `gopass recipients` will also try to load and save any missing GPG keys from and to the store.

The commands manipulating recipients, i.e. `gopass recipients add` and `gopass recipients remove` accept a `--store` flag that expects the *name of a mount point* to operate on this mounted sub store.

To check and reencrypt secrets if recipients are missing, run `gopass fsck`.

### Debugging

To debug gopass, set the environment variable `GOPASS_DEBUG_LOG` to a output filename.

### Restricting the characters in generated passwords

To restrict the characters used in generated passwords set `GOPASS_CHARACTER_SET` to any non-empty string. Please keep in mind that this can considerably weaken the strength of generated passwords.

### Using custom password generators

To use an external password generator set `GOPASS_EXTERNAL_PWGEN` to any valid executable with all required arguments. Please note that the command will be run as-is. Not parameters to control length or complexity can be passed. Any errors will be silently ignored and gopass will fall back to the internal password generator instead.

### In-place updates to existing passwords

Running `gopass [generate|insert] foo/bar` on an existing entry `foo/bar` will only update the first line of the secret, leaving any trailing data in place.

*Note: if the trailing data is marked as YAML (has a line with `---` after the password line), invalid YAML will be removed!*

### Disabling Colors

Disabling colors is as simple as setting `GOPASS_NOCOLOR` to `true`.

### Password Templates

With gopass you can create templates which are searched when executing `gopass edit` on a new secret. If the folder, or any parent folder, contains a file called `.pass-template` it's parsed as a Go template, executed with the name of the new secret and an auto-generated password and loaded into your `$EDITOR`.

This makes it easy to use templates for certain kind of secrets such as database passwords.

### Safecontent

Gopass can limit display of certain *unsafe* fields in secrets.
By default no fields are obstructed, but if the `safecontent`
config option is set to `true` the `Password` field is obstructed.
Also the special `unsafe-keys` key is evaluated. It expectes
a comma separated list of keys that will be obstructed when
printing the secret.

