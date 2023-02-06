# <img src="https://github.com/vinegar-dev/vinegar/blob/master/desktop/vinegar.svg" width="48"> Vinegar
A transparent wrapper for Roblox Player and Roblox Studio.

# Features
+ Automatic usage of the Nvidia dedicated gpu. (untested
+ Automatic applying of [RCO](https://github.com/L8X/Roblox-Client-Optimizer) FFlags, when enabled in configuration
+ Automatically fetch and install Roblox Player, Studio and rbxfpsunlocker, when needed for launch
+ Browser launch via MIME
+ Clean wine log output
+ Configuration file for setting environment variables and applying custom FFlags
+ Custom execution of wine program within wineprefix
+ Deletion of empty stderr log file
+ Fast finding of Roblox Player and Roblox Studio
+ Faster startup of rbxfpsunlocker and the Roblox Player
+ FreeBSD support
+ Logging for stderr

# TODO
+ Add watchdog for unlocker in flatpak? This needs investigation.
+ Automatically kill wineprefix when Roblox has exited
+ Better log names
+ Fetch latest version of rbxfpsunlocker
+ Old death sounds (maybe)

# Configuration
```yaml
autolaunch_rfpsu: false
use_rco_fflags: true
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
