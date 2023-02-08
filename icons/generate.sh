#!/bin/sh

for size in 16 32 48 64 128; do
	out=$size
	rm -rf $out

	size=${size}x${size}

	mkdir -p $out
	
	convert roblox-player.png -resize $size $out/com.github.vinegar.roblox.player.png
	convert roblox-studio.png -resize $size $out/com.github.vinegar.roblox.studio.png
done
