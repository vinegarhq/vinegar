# <img src="icons/48/vinegar.png"> Vinegar

![Workflow Status][workflow_img    ]
[![Version        ][version_img     ]][version     ]
[![Flathub        ][flathub_img     ]][flathub     ]
[![Report Card    ][goreportcard_img]][goreportcard]
[![Discord Server ][discord_img     ]][discord     ]

An open-source, minimal, configurable, fast bootstrapper for running Roblox on Linux.

# DISCLAIMER

Hi all, Roblox is currently blocking Wine users with the new 64-bit client for the foreseeable future. If you encounter the message "Wine is not supported", please know that it comes from Roblox's side. There's nothing we can do at the moment.

Additionally, the new version of Studio is using obscure Windows functions which are currently unsupported by Wine. Several communities are currently working to resolve this issue.

We apologize for any downtime. These updates are detrimental to Wine users. For Player, the only thing that can be done at the moment is to express feedback to Roblox, as they have mentioned they are open towards Wine usage in the future.

If you have any clue on how to continue the Roblox on Linux experience, please lend a hand!

Thank you. 

[workflow_img]: https://img.shields.io/github/actions/workflow/status/vinegarhq/vinegar/build.yml
[version]: https://github.com/vinegarhq/vinegar/releases/latest
[version_img]: https://img.shields.io/github/v/release/vinegarhq/vinegar?display_name=tag
[flathub]: https://flathub.org/apps/details/io.github.vinegarhq.Vinegar
[flathub_img]: https://img.shields.io/flathub/downloads/io.github.vinegarhq.Vinegar
[goreportcard]:     https://goreportcard.com/report/github.com/vinegarhq/vinegar
[goreportcard_img]: https://goreportcard.com/badge/github.com/vinegarhq/vinegar?style=flat-square
[discord]: https://discord.gg/dzdzZ6Pps2
[discord_img]: https://img.shields.io/discord/1069506340973707304

# Features
+ Automatic applying of [RCO](https://github.com/L8X/Roblox-Client-Optimizer) FFlags, when enabled in configuration 
  + Includes a built in FPS unlocker
  + Optimizes Roblox's performance
  + Disables a large portion of client telemetry
+ Automatic DXVK Installer and uninstaller
+ Automatic Wineprefix killer when Roblox has quit
+ Automatic Wineprefix version setter upon Wineprefix initialization
+ Browser launch via MIME
+ Custom execution of wine program within wineprefix
+ Custom launcher specified to be used when launching Roblox (eg. [GameMode](https://github.com/FeralInteractive/gamemode)).
+ Custom Wine 'root'
+ Custom Roblox Player & Studio launcher
+ Faster Multi-threaded installation and execution of Roblox
+ TOML Configuration file for setting environment variables and applying custom FFlags
+ Multiple instances of Roblox open simultaneously
+ Logging for both Vinegar and executions

# See Also
+ [Discord Server](https://discord.gg/dzdzZ6Pps2)
+ [Documentation](https://vinegarhq.github.io)
+ [Roblox-Studio-Mod-Manager](https://github.com/MaximumADHD/Roblox-Studio-Mod-Manager)
+ [Bloxstrap](https://github.com/pizzaboxer/bloxstrap)

# Acknowledgements
+ Big Thanks to [pizzaboxer](https://github.com/pizzaboxer)
+ Credits to [MaximumADHD](https://github.com/MaximumADHD)
+ Logo modified with Katie, made by the [Twemoji team](https://twemoji.twitter.com/), Licensed under [CC-BY 4.0](https://creativecommons.org/licenses/by/4.0/).
+ Katie usage authorized by [kitteh](https://ksiv.neocities.org)
