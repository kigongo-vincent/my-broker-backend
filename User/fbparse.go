package user

import (
	"fmt"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/kigongo-vincent/my-broker-backend/fbs/gen/mybroker"
)

func ParseSignInPhone(req *mybroker.RequestEnv) (string, error) {
	if req.BodyType() != mybroker.ReqPayloadSignInPhone {
		return "", fmt.Errorf("expected SignInPhone")
	}
	var t flatbuffers.Table
	if !req.Body(&t) {
		return "", fmt.Errorf("missing body")
	}
	var s mybroker.SignInPhone
	s.Init(t.Bytes, t.Pos)
	return string(s.PhoneNumber()), nil
}

func ParseVerifyOtp(req *mybroker.RequestEnv) (*mybroker.VerifyOtpBody, error) {
	if req.BodyType() != mybroker.ReqPayloadVerifyOtpBody {
		return nil, fmt.Errorf("expected VerifyOtpBody")
	}
	var t flatbuffers.Table
	if !req.Body(&t) {
		return nil, fmt.Errorf("missing body")
	}
	var v mybroker.VerifyOtpBody
	v.Init(t.Bytes, t.Pos)
	return &v, nil
}

func ParsePhonePin(req *mybroker.RequestEnv) (*mybroker.PhonePinBody, error) {
	if req.BodyType() != mybroker.ReqPayloadPhonePinBody {
		return nil, fmt.Errorf("expected PhonePinBody")
	}
	var t flatbuffers.Table
	if !req.Body(&t) {
		return nil, fmt.Errorf("missing body")
	}
	var p mybroker.PhonePinBody
	p.Init(t.Bytes, t.Pos)
	return &p, nil
}

func ParseGoogleAuth(req *mybroker.RequestEnv) (*mybroker.GoogleAuthBody, error) {
	if req.BodyType() != mybroker.ReqPayloadGoogleAuthBody {
		return nil, fmt.Errorf("expected GoogleAuthBody")
	}
	var t flatbuffers.Table
	if !req.Body(&t) {
		return nil, fmt.Errorf("missing body")
	}
	var g mybroker.GoogleAuthBody
	g.Init(t.Bytes, t.Pos)
	return &g, nil
}

func ParseUpdateProfile(req *mybroker.RequestEnv) (*mybroker.UpdateProfileBody, error) {
	if req.BodyType() != mybroker.ReqPayloadUpdateProfileBody {
		return nil, fmt.Errorf("expected UpdateProfileBody")
	}
	var t flatbuffers.Table
	if !req.Body(&t) {
		return nil, fmt.Errorf("missing body")
	}
	var x mybroker.UpdateProfileBody
	x.Init(t.Bytes, t.Pos)
	return &x, nil
}

func ParseBlock(req *mybroker.RequestEnv) (*mybroker.BlockBody, error) {
	if req.BodyType() != mybroker.ReqPayloadBlockBody {
		return nil, fmt.Errorf("expected BlockBody")
	}
	var t flatbuffers.Table
	if !req.Body(&t) {
		return nil, fmt.Errorf("missing body")
	}
	var x mybroker.BlockBody
	x.Init(t.Bytes, t.Pos)
	return &x, nil
}

func ParseReport(req *mybroker.RequestEnv) (*mybroker.ReportBody, error) {
	if req.BodyType() != mybroker.ReqPayloadReportBody {
		return nil, fmt.Errorf("expected ReportBody")
	}
	var t flatbuffers.Table
	if !req.Body(&t) {
		return nil, fmt.Errorf("missing body")
	}
	var x mybroker.ReportBody
	x.Init(t.Bytes, t.Pos)
	return &x, nil
}

func ParseRoomID(req *mybroker.RequestEnv) (*mybroker.RoomIdBody, error) {
	if req.BodyType() != mybroker.ReqPayloadRoomIdBody {
		return nil, fmt.Errorf("expected RoomIdBody")
	}
	var t flatbuffers.Table
	if !req.Body(&t) {
		return nil, fmt.Errorf("missing body")
	}
	var x mybroker.RoomIdBody
	x.Init(t.Bytes, t.Pos)
	return &x, nil
}

func ParsePostID(req *mybroker.RequestEnv) (*mybroker.PostIdBody, error) {
	if req.BodyType() != mybroker.ReqPayloadPostIdBody {
		return nil, fmt.Errorf("expected PostIdBody")
	}
	var t flatbuffers.Table
	if !req.Body(&t) {
		return nil, fmt.Errorf("missing body")
	}
	var x mybroker.PostIdBody
	x.Init(t.Bytes, t.Pos)
	return &x, nil
}

func ParseApproveUser(req *mybroker.RequestEnv) (*mybroker.ApproveUserBody, error) {
	if req.BodyType() != mybroker.ReqPayloadApproveUserBody {
		return nil, fmt.Errorf("expected ApproveUserBody")
	}
	var t flatbuffers.Table
	if !req.Body(&t) {
		return nil, fmt.Errorf("missing body")
	}
	var x mybroker.ApproveUserBody
	x.Init(t.Bytes, t.Pos)
	return &x, nil
}

func ParseUpdateID(req *mybroker.RequestEnv) (*mybroker.UpdateIdBody, error) {
	if req.BodyType() != mybroker.ReqPayloadUpdateIdBody {
		return nil, fmt.Errorf("expected UpdateIdBody")
	}
	var t flatbuffers.Table
	if !req.Body(&t) {
		return nil, fmt.Errorf("missing body")
	}
	var x mybroker.UpdateIdBody
	x.Init(t.Bytes, t.Pos)
	return &x, nil
}
