# <img src="https://github.com/vinegar-dev/vinegar/blob/master/icons/vinegar.svg" width="48"> Vinegar
A transparent wrapper for Roblox Player and Roblox Studio.

# Features
+ Automatic applying of [RCO](https://github.com/L8X/Roblox-Client-Optimizer) FFlags, when enabled in configuration
+ Automatic usage of the Nvidia dedicated gpu. (untested)
+ Automatic Wineprefix killer when Roblox Player has exited
+ Automatically fetch and install Roblox Player, Studio and rbxfpsunlocker, when needed at run time
+ Browser launch via MIME
+ Clean wine log output
+ DXVK Installer and uninstaller
+ Configuration file for setting environment variables and applying custom FFlags
+ Custom execution of wine program within wineprefix
+ Fast finding of Roblox Player and Roblox Studio
+ Faster startup of rbxfpsunlocker and the Roblox Player
+ FreeBSD support
+ Logging for stderr

# TODO
+ Add watchdog for unlocker in flatpak? This needs investigation.
+ Fetch latest version of rbxfpsunlocker
+ Handle SIGINT and SIGEXIT
+ Old death sounds (maybe)

# Configuration
The configuration file is looked at by default in `~/.config/vinegar/config.yaml`.
```yaml
rfpsu: false
rco: true
fflags:
  FFlagFoo: "null"
  FFlagBar: true
  FFlagBaz: 2147483648
env:
  foo: bar
```
By default, [RCO](https://github.com/L8X/Roblox-Client-Optimizer) FFlags will be installed automatically, and rbxfpsunlocker will be disabled by default.

# Why RCO?
Credits to [L8X](https://github.com/L8X), RCO's FFlags optimizes Roblox's performance, caching, and textures whilst removing the FPS Unlock without the need for rbxfpsunlocker.

The Discord server for Vinegar can be found [here](https://discord.gg/dzdzZ6Pps2).
