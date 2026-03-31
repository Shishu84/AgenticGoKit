package main

import (
    "fmt"
    "github.com/Shishu84/AgenticGoKit/mywork"
)

func main() {
    m := mywork.Memory{}
    m.Remember("Week 2 Git workflow executed")
    m.Remember("Implemented agent memory module")

    fmt.Println("Memory Recall:", m.Recall())

    m.Forget()
    fmt.Println("After Forget:", m.Recall())
}
