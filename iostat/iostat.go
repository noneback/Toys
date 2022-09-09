package iostat

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"
)

type ExtIostatAttr struct {
	iopsR         float64
	iopsW         float64
	ioThroughOutR float64
	ioThroughOutW float64
	ioUtil        float64
	ioAwaitR      float64
	ioAwaitW      float64
	ioAvgRqSzR    float64
	ioAvgRqSzW    float64
	ioAvgRqSize   float64
	ioAvgQuSize   float64
}

type innerIOAttr struct {
	major uint64
	minor uint64
	dev   string

	/* of sectors read */
	rdSectors uint64
	/* of sectors written */
	wrSectors uint64
	/* of sectors discarded */
	dcSectors uint64
	/* of read operations issued to the device */
	rdIOs uint64
	/* of read requests merged */
	rdMerges uint64
	/* of write operations issued to the device */
	wrIOs uint64
	/* of write requests merged */
	wrMerges uint64
	/* of discard operations issued to the device */
	dcIOs uint64
	/* of discard requests merged */
	dcMerges uint64
	/* of flush requests issued to the device */
	flIOs uint64
	/* Time of read requests in queue */
	rdTicks uint64
	/* Time of write requests in queue */
	wrTicks uint64
	/* Time of discard requests in queue */
	dcTicks uint64
	/* Time of flush requests in queue */
	flTicks uint64
	/* of I/Os in progress */
	iosPgr uint64
	/* of ticks total (for this device) for I/O */
	totTicks uint64
	/* of ticks requests spent in queue */
	rqTicks uint64
}

type IostatCollector struct {
	diskstatFd      *os.File
	interval        uint32
	firstCollect    bool
	ext             ExtIostatAttr
	dev             string
	cur, pre        *innerIOAttr
	lastCollectTime time.Time
	sectorSize      uint32
	done            chan struct{}
}

type Config struct {
	internal   uint32
	dev        string
	sectorSize uint32 // in byte
}

func NewIostatCollector(opt Config) *IostatCollector {
	return &IostatCollector{
		interval:     opt.internal,
		dev:          opt.dev,
		firstCollect: true,
		sectorSize:   opt.sectorSize,
		done:         make(chan struct{}),
	}
}

func (ic *IostatCollector) Init() {
	if ic.diskstatFd == nil {
		var err error
		ic.diskstatFd, err = os.Open("/proc/diskstats")
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func (ic *IostatCollector) Start() {
	log.Println("Run on background")
	go ic.run()

}

func (ic *IostatCollector) run() {
	ticker := time.NewTicker(time.Duration(ic.interval) * time.Second)
	for {
		go ic.updateIOStat()
		select {
		case <-ticker.C:
		case <-ic.done:
			ticker.Stop()
			return
		}
	}

}

func (ic *IostatCollector) updateIOStat() {
	scanner := bufio.NewScanner(ic.diskstatFd)
	for scanner.Scan() {
		line := scanner.Text()
		innerIOAttrs, err := ic.parseLine(line)
		if err != nil {
			log.Fatalln(err)
		}
		if innerIOAttrs.dev != ic.dev {
			continue
		}
		ic.pre = ic.cur
		ic.cur = innerIOAttrs
		ic.lastCollectTime = time.Now()
		if !ic.firstCollect {
			// calculate ExtIoStatAttr
			ic.computeExtAttrs()
			fmt.Println(ic.ext.iopsR)
			// fmt.Printf("%v: %+v, %v\n", innerIOAttrs.dev, ic.ext.iopsR, ic.cur.rdIOs)
			// fmt.Println(line)
		}

	}
	ic.firstCollect = false
	if _, err := ic.diskstatFd.Seek(0, 0); err != nil {
		log.Fatalln(err)
	}
}

func (ic *IostatCollector) computeExtAttrs() {
	ic.ext.ioAwaitR = 0
	ic.ext.ioAwaitW = 0
	ic.ext.ioAvgRqSzR = 0
	ic.ext.ioAvgRqSzW = 0
	fmt.Printf("debug: pre %v cur %v,%v %v %v\n", ic.pre.rdIOs, ic.cur.rdIOs, ic.cur.rdIOs-ic.pre.rdIOs, ic.interval, float64(ic.cur.rdIOs-ic.pre.rdIOs)/float64(ic.interval))
	ic.ext.iopsR = float64(ic.cur.rdIOs-ic.pre.rdIOs) / float64(ic.interval)
	ic.ext.iopsW = float64(ic.cur.wrIOs-ic.pre.wrIOs) / float64(ic.interval)

	ic.ext.ioThroughOutR = float64((ic.cur.rdSectors-ic.pre.rdSectors)*uint64(ic.sectorSize)) / float64(ic.interval) / 1024.0 / 1024.0 // in MB
	ic.ext.ioThroughOutW = float64((ic.cur.wrSectors-ic.pre.wrSectors)*uint64(ic.sectorSize)) / float64(ic.interval) / 1024.0 / 1024.0 // in MB

	if (ic.cur.rdIOs - ic.pre.rdIOs) != 0 {
		ic.ext.ioAwaitR = float64(ic.cur.rdTicks-ic.pre.rdTicks) / float64(ic.cur.rdIOs-ic.pre.rdIOs)
		ic.ext.ioAvgRqSzR = float64(ic.cur.rdSectors-ic.pre.rdSectors) / float64(ic.cur.rdIOs-ic.pre.rdIOs)
	}
	if (ic.cur.rdIOs - ic.pre.rdIOs) != 0 {
		ic.ext.ioAwaitW = float64(ic.cur.wrTicks-ic.pre.wrTicks) / float64(ic.cur.wrIOs-ic.pre.wrIOs)
		ic.ext.ioAvgRqSzW = float64(ic.cur.wrSectors-ic.pre.wrSectors) / float64(ic.cur.wrIOs-ic.pre.wrIOs)
	}

	ic.ext.ioUtil = float64(ic.cur.totTicks-ic.pre.totTicks) / float64(ic.interval)
	ic.ext.ioAvgQuSize = float64(ic.cur.rqTicks-ic.pre.rqTicks) / float64(ic.interval) / 1000.0
}

func (ic *IostatCollector) parseLine(line string) (*innerIOAttr, error) {
	var iosPgr, totTicks, rqTicks, wrTicks, dcTicks, flTicks,
		rdIOs, rdMergesOrRdSec, rdTicksOrWrSec, wrIOs,
		wrMerges, rdSecOrWrIOs, wrSec,
		dcIOs, dcMerges, dcSec, flIOs, major, minor uint64
	var devName string
	innerAttr := &innerIOAttr{}

	n, err := fmt.Sscanf(line, "%v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v",
		&major, &minor, &devName,
		&rdIOs, &rdMergesOrRdSec, &rdSecOrWrIOs, &rdTicksOrWrSec,
		&wrIOs, &wrMerges, &wrSec, &wrTicks, &iosPgr, &totTicks, &rqTicks,
		&dcIOs, &dcMerges, &dcSec, &dcTicks,
		&flIOs, &flTicks)
	if err != nil {
		return nil, err
	}
	innerAttr.major = major
	innerAttr.minor = minor
	innerAttr.dev = devName

	if n >= 14 {
		innerAttr.rdIOs = rdIOs
		innerAttr.rdMerges = rdMergesOrRdSec
		innerAttr.rdSectors = rdSecOrWrIOs
		innerAttr.rdTicks = rdTicksOrWrSec
		innerAttr.wrIOs = wrIOs
		innerAttr.wrMerges = wrMerges
		innerAttr.wrSectors = wrSec
		innerAttr.wrTicks = wrTicks
		innerAttr.iosPgr = iosPgr
		innerAttr.totTicks = totTicks
		innerAttr.rqTicks = rqTicks
		if n >= 18 {
			innerAttr.dcIOs = dcIOs
			innerAttr.dcMerges = dcMerges
			innerAttr.dcSectors = dcSec
			innerAttr.dcTicks = dcTicks
			if n >= 20 {
				innerAttr.flIOs = flIOs
				innerAttr.flTicks = flTicks
			}
		}
	}
	return innerAttr, nil
}

func (ic *IostatCollector) Stop() {
	fmt.Println("Stop Background Task")
	ic.done <- struct{}{}

	if ic.diskstatFd != nil {
		ic.diskstatFd.Close()
	}
}
