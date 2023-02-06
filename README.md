# <img src="https://github.com/vinegar-dev/vinegar/blob/master/desktop/vinegar.svg" width="48"> Vinegar
A transparent wrapper for Roblox Player and Roblox Studio.

# Features
+ Configuration file for setting environment variables and applying custom FFlags
+ Logging for stderr
+ Handling arguments parsing and forwarding of RobloxPlayerLauncher (to be used)
+ FreeBSD support
+ Custom execution of wine program within wineprefix
+ Fast finding of Roblox Player and Roblox Studio
+ Clean wine log output
+ Automatic applying of [RCO](https://github.com/L8X/Roblox-Client-Optimizer) FFlags
+ (Untested) Automatic usage of the Nvidia dedicated gpu.
+ Deletion of empty log files
+ Sets up a Wine prefix automatically
+ Automatically fetch and install Roblox Player, Studio and rbxfpsunlocker
+ Browser launch (testing)
+ Faster startup of rbxfpsunlocker and the Roblox Player

# TODO
+ FSYNC/ESYNC toggles
+ Old death sounds (maybe)
+ Simple graphical user interface for easy modification of the configuration, or to launch Wine apps
+ Fetch latest version of Roblox, when RobloxPlayerLauncher is not used.
+ Better log names
+ Fetch latest version of rbxfpsunlocker
+ Add watchdog for unlocker in flatpak? This needs investigation.
+ Automatically kill wineprefix when Roblox has exited
+ Add installation failure detection

# Configuration
```yaml
use_rco_fflags: true
fflags:
  FFlagFoo: "null"
  FFlagBar: true
  FFlagBaz: 2147483648
env:
  foo: bar
```
By default, [RCO](https://github.com/L8X/Roblox-Client-Optimizer) FFlags will be installed automatically.
# Why RCO?
Credits to [L8X](https://github.com/L8X), RCO's FFlags optimizes Roblox's performance, caching, and textures whilst removing the FPS Unlock without the need for rbxfpsunlocker.

The Discord server for Vinegar can be found [here](https://discord.gg/dzdzZ6Pps2).
