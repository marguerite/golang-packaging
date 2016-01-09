# golang-packaging

RPM macros and utilities for golang packaging.

A golang packager can just say

>  BuildRequires: golang-packaging 

and let the included scripts handle Provides/Requires for
you automatically. You can also say 

>  BuildRequires: golang(xxx) 

in specifications for packages built with golang-packaging.

# Macros explained

https://en.opensuse.org/openSUSE:Packaging_Go
