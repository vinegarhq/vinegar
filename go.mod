module github.com/vinegarhq/vinegar

go 1.20

require (
	github.com/BurntSushi/toml v1.3.2
	github.com/adrg/xdg v0.4.0
	github.com/otiai10/copy v1.12.0
	golang.org/x/sync v0.3.0
)

require (
	dario.cat/mergo v1.0.0
	gioui.org v0.3.0
	golang.org/x/sys v0.11.0
)

require (
	gioui.org/cpu v0.0.0-20210817075930-8d6a761490d2 // indirect
	gioui.org/shader v1.0.6 // indirect
	github.com/go-text/typesetting v0.0.0-20230803102845-24e03d8b5372 // indirect
	golang.org/x/exp v0.0.0-20221012211006-4de253d81b95 // indirect
	golang.org/x/exp/shiny v0.0.0-20220827204233-334a2380cb91 // indirect
	golang.org/x/image v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
)

retract (
	[v1.0.0, v1.1.3]
	v0.0.1
)
