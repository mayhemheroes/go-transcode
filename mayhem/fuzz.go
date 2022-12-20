package fuzz

import "strconv"
import "github.com/m1k1o/go-transcode/hlsproxy"

func mayhemit(bytes []byte) int {

    var num int
    if len(bytes) > 2 {
        num, _ = strconv.Atoi(string(bytes[0]))
        bytes = bytes[1:]

        switch num {
    
        case 0:
            content := string(bytes)
            part1 := content[0:len(content)/2]
            part2 := content[len(content)/2:len(content)]

            hlsproxy.New(part1, part2)
            return 0

        default:
            content := string(bytes)
            part1 := content[0:len(content)/3]
            part2 := content[len(content)/3:2*(len(content)/3)]
            part3 := content[2*(len(content)/3):len(content)]

            hlsproxy.RelativePath(part1, part2, part3)
            return 0

        }
    }
    return 0
}

func Fuzz(data []byte) int {
    _ = mayhemit(data)
    return 0
}