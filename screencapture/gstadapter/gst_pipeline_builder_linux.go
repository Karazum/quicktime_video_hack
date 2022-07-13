//go:build linux
// +build linux

package gstadapter

import (
	"fmt"
	"github.com/danielpaulus/gst"
	"github.com/lijo-jose/glib"
	"os/exec"
	"regexp"
	"strconv"
)

func setupLivePlayAudio(pl *gst.Pipeline) {

	/*hack: I do not know why, but audio on my linux box wont play when using a simple wavpars.
	On MAC OS it works without any problems though. A hacky workaround to get audio playing that I came up with was
	to encode audio into ogg/vorbis and directly decode it again.
	*/

	vorbisenc := gst.ElementFactoryMake("vorbisenc", "vorbisenc_01")
	checkElem(vorbisenc, "vorbisenc_01")

	oggmux := gst.ElementFactoryMake("oggmux", "oggmux_01")
	checkElem(oggmux, "oggmux_01")

	oggdemux := gst.ElementFactoryMake("oggdemux", "oggdemux")
	checkElem(oggdemux, "oggdemux")

	vorbisdec := gst.ElementFactoryMake("vorbisdec", "vorbisdec")
	checkElem(vorbisdec, "vorbisdec")

	audioconvert2 := gst.ElementFactoryMake("audioconvert", "audioconvert_02")
	checkElem(audioconvert2, "audioconvert_02")

	//endhack

	autoaudiosink := gst.ElementFactoryMake("autoaudiosink", "autoaudiosink_01")
	checkElem(autoaudiosink, "autoaudiosink_01")
	autoaudiosink.SetProperty("sync", false)

	pl.Add(vorbisenc, oggmux, oggdemux, vorbisdec, audioconvert2, autoaudiosink)
	pl.GetByName("queue2").Link(vorbisenc)

	vorbisenc.Link(vorbisdec)
	vorbisdec.Link(audioconvert2)

	audioconvert2.Link(autoaudiosink)

}

func setUpVideoPipeline(pl *gst.Pipeline) *gst.AppSrc {
	asrc := gst.NewAppSrc("my-video-src")
	asrc.SetProperty("is-live", true)

	queue1 := gst.ElementFactoryMake("queue", "queue_11")
	checkElem(queue1, "queue_11")

	h264parse := gst.ElementFactoryMake("h264parse", "h264parse_01")
	checkElem(h264parse, "h264parse")

	avdec_h264 := gst.ElementFactoryMake("avdec_h264", "avdec_h264_01")
	checkElem(avdec_h264, "avdec_h264_01")

	queue2 := gst.ElementFactoryMake("queue", "queue_12")
	checkElem(queue2, "queue_12")

	videoconvert := gst.ElementFactoryMake("videoconvert", "videoconvert_01")
	checkElem(videoconvert, "videoconvert_01")

	videoscale := gst.ElementFactoryMake("videoscale", "videoscale_01")
	checkElem(videoscale, "videoscale_01")

	//video := gst.ElementFactoryMake("video/x-raw, width=500, height=1000", "video")
	//checkElem(video, "video")

	queue3 := gst.ElementFactoryMake("queue", "queue_13")
	checkElem(queue3, "queue_13")

	sink := gst.ElementFactoryMake("ximagesink", "ximagesink_01")
	checkElem(sink, "ximagesink01")
	//sink.SetProperty("display", "localhost:201")
	sink.SetProperty("sync", false) //see gst_adapter_macos comment

	pl.Add(asrc.AsElement(), queue1, h264parse, avdec_h264, queue2, videoconvert, videoscale, queue3, sink)

	asrc.Link(queue1)
	queue1.Link(h264parse)
	h264parse.Link(avdec_h264)
	avdec_h264.Link(queue2)
	queue2.Link(videoconvert)
	videoconvert.Link(videoscale)
	//videoscale.Link(queue3)
	//video.Link(queue3)
	queue3.Link(sink)
	out, err := exec.Command("xdpyinfo").Output()
	if err != nil {
		fmt.Printf("%s", err)
	}
	outputString := string(out)
	r, _ := regexp.Compile("dimensions: *(\\d+)x(\\d+)")
	res := r.FindAllStringSubmatch(outputString, -1)
	width, _ := strconv.Atoi(res[0][1])
	height, _ := strconv.Atoi(res[0][2])

	filter := gst.NewCapsSimple(
		"video/x-raw",
		glib.Params{
			"width":  int32(width),
			"height": int32(height),
		},
	)
	videoscale.LinkFiltered(queue3, filter)
	return asrc
}
