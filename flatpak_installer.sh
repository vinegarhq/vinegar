#!/bin/bash
echo "Roblox Installer by blan.ks and proofread with cleaned up code by killertofus"
echo "Before the installation continues I heavily encourage you to read through the script and ensure it is not malicious"
sleep 5

if grep avx /proc/cpuinfo 1>/dev/null
        then
        echo "you computer meets the requirements to play roblox"
        else
echo "Sadly your computer is too old to play ROBLOX! With the release of Hyperion/Byfron anticheat brought requirements that your computer does not meet. (Missing AVX)"
fi
package='flatpak'
which flatpak >/dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "flatpak is installed."
else
    echo "flatpak is missing and is required for this script. installing it for you."
fi
  if [ -x "$(command -v apk)" ];       then sudo apk add --no-cache $package
  elif [ -x "$(command -v apt)" ]; then sudo apt install $package
  elif [ -x "$(command -v dnf)" ];     then sudo dnf install $package
  elif [ -x "$(command -v zypper)" ];  then sudo zypper install $package
  elif [ -x "$(command -v pacman)" ]; then sudo pacman -S $package
  else echo "FAILED TO INSTALL PACKAGE: Package manager not found. You must manually install: $package">&2;fi

echo "Setting up flatpak"
flatpak --user remote-add --if-not-exists flathub https://dl.flathub.org/repo/flathub.flatpakrepo
 
#Flatpak is used as that is easiest
echo "Installing Vinegar flatpak"
echo "Please press y at the prompt. Vinegar is the thing that lets you play roblox"
flatpak install io.gitub.vinegarhq.Vinegar
 
flatpak list --app | grep "io.github.vinegarhq.Vinegar" || \
echo "It seems Vinegar was not installed. Please run again and make sure you pressed y" && exit
 
echo "Installing WineGE"
echo "This may take a minute depending on your wifi and computer speeds"
cd ~/


curl -s https://api.github.com/repos/GloriousEggroll/wine-ge-custom/releases/latest \
| grep "wine-lutris-GE-Proton.*tar.xz" \
| cut -d : -f 2,3 \
| tr -d \" \
| wget -qi -
tar xzf *
mv ~/.local/share/lutris/runners/wine/

echo "Setting up config.toml"


mkdir -p ~/.var/app/io.github.vinegarhq.Vinegar/config/vinegar
if test -f ~/.var/app/io.github.vinegarhq.Vinegar/config/vinegar/config.toml; then
    echo "User has already made config.toml. changed to config.toml.bak"
    mv ~/.var/app/io.github.vinegarhq.Vinegar/config/vinegar/config.toml ~/.var/app/io.github.vinegarhq.Vinegar/config/vinegar/config.toml.bak
fi
touch ~/.var/app/io.github.vinegarhq.Vinegar/config/vinegar/config.toml
echo -e "[global]\nwineroot = \"$HOME/wine-lutris-GE-Proton8-21-x86_64\"\nchannel = \"zavatarteam2\"" > ~/.var/app/io.github.vinegarhq.Vinegar/config/vinegar/config.toml

echo -e "Launching ROBLOX..."
flatpak run io.github.vinegarhq.Vinegar player
echo "If ROBLOX does not launch after a couple minutes or performs badly while in a game please press enter"
echo "Otherwise you may exit out of this window now and play roblox"
echo "NOTE: If logging in from the client does not work please do it from the website and start a game from there"
end 1

