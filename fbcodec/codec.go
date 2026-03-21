package fbcodec

import (
	"github.com/gofiber/fiber/v2"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/kigongo-vincent/my-broker-backend/fbs/gen/mybroker"
)

const ContentType = "application/x-flatbuffer"

func Send(c *fiber.Ctx, httpStatus int, body []byte) error {
	c.Status(httpStatus)
	c.Set("Content-Type", ContentType)
	return c.Send(body)
}

// FinishEnv completes Envelope as root. String offsets must already be created on b; payloadOff must be built before this call.
func FinishEnv(b *flatbuffers.Builder, httpStatus int32, msgOff, tokOff, warOff flatbuffers.UOffsetT, otp int32, pt mybroker.ApiPayload, payloadOff flatbuffers.UOffsetT) []byte {
	mybroker.EnvelopeStart(b)
	mybroker.EnvelopeAddPayload(b, payloadOff)
	mybroker.EnvelopeAddPayloadType(b, pt)
	mybroker.EnvelopeAddWarning(b, warOff)
	mybroker.EnvelopeAddOtp(b, otp)
	mybroker.EnvelopeAddToken(b, tokOff)
	mybroker.EnvelopeAddMsg(b, msgOff)
	mybroker.EnvelopeAddHttpStatus(b, httpStatus)
	root := mybroker.EnvelopeEnd(b)
	b.Finish(root)
	return b.FinishedBytes()
}

func SendEmpty(c *fiber.Ctx, httpStatus int, msg string) error {
	b := flatbuffers.NewBuilder(256)
	msgOff := b.CreateString(msg)
	mybroker.EmptyStart(b)
	empty := mybroker.EmptyEnd(b)
	raw := FinishEnv(b, int32(httpStatus), msgOff, 0, 0, 0, mybroker.ApiPayloadEmpty, empty)
	return Send(c, httpStatus, raw)
}

func SendError(c *fiber.Ctx, httpStatus int, msg string) error {
	return SendEmpty(c, httpStatus, msg)
}

// BuildAndSend finishes an Envelope after build() returns the payload table offset (same builder).
func BuildAndSend(c *fiber.Ctx, httpStatus int, msg, token, warning string, otp int32, pt mybroker.ApiPayload, initCap int, build func(*flatbuffers.Builder) flatbuffers.UOffsetT) error {
	b := flatbuffers.NewBuilder(initCap)
	payload := build(b)
	msgOff := b.CreateString(msg)
	var tokOff, warOff flatbuffers.UOffsetT
	if token != "" {
		tokOff = b.CreateString(token)
	}
	if warning != "" {
		warOff = b.CreateString(warning)
	}
	raw := FinishEnv(b, int32(httpStatus), msgOff, tokOff, warOff, otp, pt, payload)
	return Send(c, httpStatus, raw)
}
