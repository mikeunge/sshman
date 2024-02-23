# SSHMan - SSH management done right

> Manage your ssh profiles with ease.

## About

SSHMan is a CLI tool for managing ssh keys & connection strings.

### But why?

We have all been there, typing ```ssh -i``` and pressing the UP arrow till our fingers hurt to find the right connection string for the server we want to connect to.

This tool is here to solve this.

### Commands

```bash
Available Options:

  -h  --help              Print help information.
  -v  --version           Prints the version.
      --about             Prints information about the program.
  -l  --list              List of all available SSH profiles.
  -c  --connect           Connect to a server with profile.
  -n  --new               Create a new SSH profile.
  -u  --update            Update an SSH profile.
  -d  --delete            Delete SSH profiles.
  -e  --export            Export profiles (for eg. sharing).
```

## Todo

- [ ] Functional CLI
  - [ ] Verify / Validate the ssh keyfile
- [ ] Store & retrieve SSH keys
  - [x] Get user & ip/hostnames
  - [ ] Conenct directly to server (_if wanted_)
- [ ] Possibility to encrypt the SSH keys in the database
- [ ] Export individual keys & user information for sharing
