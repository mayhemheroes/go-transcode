package fuzzHlsproxy

import "strconv"
import "github.com/m1k1o/go-transcode/hlsproxy"
import fuzz "github.com/AdaLogics/go-fuzz-headers"

func mayhemit(bytes []byte) int {

    var num int
    if len(bytes) > 2 {
        num, _ = strconv.Atoi(string(bytes[0]))
        bytes = bytes[1:]

        switch num {
    
        case 0:
            fuzzConsumer := fuzz.NewConsumer(bytes)
            var part1 string
            var part2 string
            err := fuzzConsumer.CreateSlice(&part1)
            if err != nil {
                return 0
            }

            err = fuzzConsumer.CreateSlice(&part2)
            if err != nil {
                return 0
            }

            hlsproxy.New(part1, part2)
            return 0

        default:
            fuzzConsumer := fuzz.NewConsumer(bytes)
            var part1 string
            var part2 string
            var part3 string
            err := fuzzConsumer.CreateSlice(&part1)
            if err != nil {
                return 0
            }

            err = fuzzConsumer.CreateSlice(&part2)
            if err != nil {
                return 0
            }

            err = fuzzConsumer.CreateSlice(&part3)
            if err != nil {
                return 0
            }


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