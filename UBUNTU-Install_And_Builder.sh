#!/bin/bash

if [[ $EUID -ne 0 ]]; then
    echo "Please run this script as root."
    exit 1
fi

# copy pasted from a stackoverflow post to be honest.
exitfn () {
    trap SIGINT     
    echo ""; # whilst I could just do echo; I assume there may be bash noobs whom try to read this. So for readablitity and understanding I did it like this.       
    echo "We recommend you use the flatpak install. Vinegar can be installed easily via flatpak." 
    exit
}

trap "exitfn" INT # implement exitfn function to hook interuptions

# Main message that the user sees

echo ""
echo ""
echo "Please consult https://vinegarhq.org/Vinegar/Installation/guides/source.html before trying this"
echo "Also please visit: https://www.gtk.org/ and install the latest version of GTK4 or GTK4 4.18>"
echo ""
echo ""
echo "You are building and going to run an automated script, it is important that you understand that just because it is from a trustworthy repository that it still might not be safe."
echo "You are running this with root/sudo (super user privs) which means we could do whatever we want."
echo "------------------------------------------------------------"
echo "What we are installing: golang, make, git, libgtk4-1"
echo "We will automatically install before compiling and cloning the repository for vinegar. If you are running this inside of the repo of vinegar, it will be cloned again. (ex: ./vinegar) "
echo "-------------------------------------------------------------"
read -p "If you understand, accept the risk and want to proceed, please press enter otherwise please exit the bash script by doing: CTRL + C" </dev/tty

# echo "You have chosen to accept the risks. We will be continuing." # gets cleared, not really needed either, was here for debugging mostly when testing trap implementation
clear

echo "We are now installing required packages to your system (golang, make, git, libgtk4-1)"

# update all avalible package info, then we will force the confirm on installation to ensure and assume the packages are installed for later.
sudo apt-get update
sudo apt-get install make git golang gettext libgtk-4-1 -y

# gettext libgtk-4-1
# One is to install GTK4 and the other is for make. 

clear
echo "We are now going to clone the vinegar repository and build."
echo ""
echo "Cloning repository."

git clone https://github.com/vinegarhq/vinegar.git
cd vinegar
make install
make

chmod +x "vinegar" # allow the binary to be executable

# Confirm the file has been created and is executable
if [[ -x "vinegar" ]]; then
    echo "Build successful, vinegar binary made."
    echo "Please NOTE: If you are not running an version of GTK4 < 4.18 please visit https://www.gtk.org/ and download a newer version if not 4.18 to ensure little dependency issues."
    echo "Additional link for linux users: https://www.gtk.org/docs/installations/linux/"
    exit
else
    echo "Failed to build, please refer to the docs: https://vinegarhq.org/Vinegar/Installation/guides/source.html"
    exit
fi
