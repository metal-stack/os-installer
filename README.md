# os-installer

The OS installer is used to configure a machine according to the given allocation specification, like configuring:

- Network interfaces
- FRR configuration
- Metal user and authorized keys
- Bootloader configuration
- Ignition and cloud-init userdata
- ...

It currently supports the officially published operating system images from the [metal-images repository](https://github.com/metal-stack/metal-images).

The installer is executed by the [metal-hammer](https://github.com/metal-stack/metal-hammer) in a chroot, pointing to the root of the uncompressed operating system image.

The input configuration for the installer are:

- The `MachineDetails` as defined in [api/v1/api.go](./api/v1/api.go)
- The `MachineAllocation` as defined in the [API repository](https://github.com/metal-stack/api/)
- An optional installer `Config` as defined in [api/v1/api.go](./api/v1/api.go) (for building own images)
