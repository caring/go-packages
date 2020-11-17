# Caring Go Packages
This repo contains reusable packages to use in caring.com's go projects.


## How to Use
1. Include this package into your GOPRIVATE variable `export GOPRIVATE="github.com/caring/go-packages"`
2. Run "go get github.com/caring/go-packages/pkg/[package you want]" or import into your go mod project.
3. Begin using in your project.



## Contributing
If you want to add to the go packages project, follow the process below.

1. Add new package directory into the pkg/[package you want directory]
    - Do not put packages directly into the pkg folder. Every package needs its own subfolder.
    - Be more descriptive and less creative with package names. (see: https://blog.golang.org/package-names)
    - Document all package names, types & functions. Add implementation notes for public content.
2. Add examples into examples directory. 
    - Follow same package / directory pattern used in pkg directory.
    - Include enough example content to cover all implementations.
