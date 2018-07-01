package reporter

import (
    "github.com/tobyhede/go-underscore"
	log "github.com/sirupsen/logrus"
)

func SumZeroes(numbers []int) (int) {
    var sum int

    fn := func(v, i int) {
        sum += v
    }
    un.EachInt(fn, numbers)
    log.Infof("%#v", sum) //15
    return sum
}

func BuildReport() {

}
