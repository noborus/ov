/*
Package oviewer provides a pager for terminals.

oviewer displays the contents of the file or reader etc...
on the current terminal screen.
After running a Run, oveiwer does not return until the user finishes running.
So if you want to do something in concurrent, you need to use goroutine.

There is also a simple usage example below:
https://github.com/noborus/mdviewer/

  package main

  import (
      "github.com/noborus/ov/oviewer"
  )

  func main() {
      ov, err := oviewer.Open("main.go")
      if err != nil {
        panic(err)
      }
      if err := ov.Run(); err != nil {
        panic(err)
      }
  }
*/
package oviewer
