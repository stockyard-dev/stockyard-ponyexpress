package server

import "github.com/stockyard-dev/stockyard-ponyexpress/internal/license"

type Limits struct { MaxEmailsDay int; Templates bool }
var freeLimits = Limits{MaxEmailsDay: 100, Templates: true}
var proLimits = Limits{MaxEmailsDay: 0, Templates: true}
func LimitsFor(info *license.Info) Limits { if info != nil && info.IsPro() { return proLimits }; return freeLimits }
