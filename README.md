# <img src="internal/splash/vinegar.png"> Vinegar

![Workflow Status][workflow_img    ]
[![Version        ][version_img     ]][version     ]
[![Flathub        ][flathub_img     ]][flathub     ]
[![Report Card    ][goreportcard_img]][goreportcard]
[![Discord Server ][discord_img     ]][discord     ]

An open-source, minimal, configurable, fast bootstrapper for running Roblox on Linux.

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
+ Automatic DXVK Installer and uninstaller
+ Automatic Wineprefix killer when Roblox has quit
+ Automatic removal of outdated cached packages and versions of Roblox
+ Discord Rich Presence support
+ Roblox's logs appear within Vinegar
+ FPS Unlocking for Player by default, without rbxfpsunlocker
+ Browser launch via MIME
+ Custom execution of wine program within wineprefix
+ TOML configuration file
  + Force a specific version of Roblox
  + Select/Force Roblox release channels, lets you opt into non-production Roblox release channels
  + Custom launcher specified to be used when launching Roblox (eg. [GameMode](https://github.com/FeralInteractive/gamemode)).
  + Wine Root, allows you to set a specific wine installation path
  + Sanitization of environment
  + Set different environment variables and FFlags for both Player and Studio
+ Modifications of Roblox via the Overlay directory, overwriting Roblox's files; such as re-adding the old death sound
+ Faster Multi-threaded installation and extraction of Roblox
+ Multiple instances of Roblox open simultaneously
+ Loading window during setup
+ Logging for both Vinegar and Wine

# See Also
+ [Discord Server](https://discord.gg/dzdzZ6Pps2)
+ [Documentation](https://vinegarhq.github.io)
+ [Roblox-Studio-Mod-Manager](https://github.com/MaximumADHD/Roblox-Studio-Mod-Manager)
+ [Bloxstrap](https://github.com/pizzaboxer/bloxstrap)

# Acknowledgements
+ Credits to
  + [pizzaboxer](https://github.com/pizzaboxer)
  + [MaximumADHD](https://github.com/MaximumADHD)
+ Logo modified with Katie, made by the [Twemoji team](https://twemoji.twitter.com/), Licensed under [CC-BY 4.0](https://creativecommons.org/licenses/by/4.0/).
+ Katie usage authorized by [karliflux](https://karliflux.neocities.org)
