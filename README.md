# go-mpv

Go bindings for [libmpv](https://mpv.io/).

> Why reinventing the wheel?
> 
> Because the old wheel didn't work for me
>
> And enhance memory safety

----

## install

linux
```
sudo apt install libmpv-dev
go get github.com/aynakeya/go-mpv
```


macos
```
brew install mpv
go get github.com/aynakeya/go-mpv
```

windows
1. compile or download mpv-2.dll.
2. copy mpv-2.dll and mpv folder to system path
```
go get github.com/aynakeya/go-mpv
```

# Known bugs

- SetProperty/SetOption doesn't work for FORMAT_STRING. Using Node or SetPropertyString instead.