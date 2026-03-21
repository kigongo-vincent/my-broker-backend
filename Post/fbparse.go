package post

import (
	"encoding/json"
	"fmt"

	flatbuffers "github.com/google/flatbuffers/go"
	usr "github.com/kigongo-vincent/my-broker-backend/User"
	"github.com/kigongo-vincent/my-broker-backend/fbs/gen/mybroker"
	"gorm.io/datatypes"
)

func ParseCreatePostBody(req *mybroker.RequestEnv) (*mybroker.CreatePostBody, error) {
	if req.BodyType() != mybroker.ReqPayloadCreatePostBody {
		return nil, fmt.Errorf("expected CreatePostBody")
	}
	var t flatbuffers.Table
	if !req.Body(&t) {
		return nil, fmt.Errorf("missing body")
	}
	var body mybroker.CreatePostBody
	body.Init(t.Bytes, t.Pos)
	return &body, nil
}

func PostFromCreateBody(body *mybroker.CreatePostBody) (usr.Post, error) {
	var p usr.Post
	p.PayWaterBills = body.PayWaterBills()
	p.PayElectricityBills = body.PayElectricityBills()
	p.PayForTrash = body.PayForTrash()
	p.HasParking = body.HasParking()
	p.RequiredFirstMonthsPaid = int(body.RequiredFirstMonthsPaid())
	p.Units = string(body.Units())
	p.IsNegotiable = body.IsNegotiable()
	p.ReviewDisclaimerAgreed = body.ReviewDisclaimerAgreed()
	p.Type = string(body.Type())
	p.Bedrooms = string(body.Bedrooms())
	p.Bathrooms = string(body.Bathrooms())
	p.Toilets = string(body.Toilets())
	if pr := body.Price(nil); pr != nil {
		p.Price.Currency = string(pr.Currency())
		p.Price.Amount = int(pr.Amount())
	}
	if loc := body.Location(nil); loc != nil {
		p.Location.Lat = loc.Lat()
		p.Location.Lon = loc.Lon()
		p.Location.Name = string(loc.Name())
	}
	n := body.ImagesLength()
	imgs := make([]string, 0, n)
	for i := 0; i < n; i++ {
		if v := body.Images(i); v != nil {
			imgs = append(imgs, string(v))
		}
	}
	if len(imgs) > 0 {
		raw, err := json.Marshal(imgs)
		if err != nil {
			return p, err
		}
		p.Images = datatypes.JSON(raw)
	}
	amen := string(body.Amenities())
	if amen != "" {
		p.Ammenities = datatypes.JSON(amen)
	}
	return p, nil
}

func ParseApprovePost(req *mybroker.RequestEnv) (*mybroker.ApprovePostBody, error) {
	if req.BodyType() != mybroker.ReqPayloadApprovePostBody {
		return nil, fmt.Errorf("expected ApprovePostBody")
	}
	var t flatbuffers.Table
	if !req.Body(&t) {
		return nil, fmt.Errorf("missing body")
	}
	var x mybroker.ApprovePostBody
	x.Init(t.Bytes, t.Pos)
	return &x, nil
}
