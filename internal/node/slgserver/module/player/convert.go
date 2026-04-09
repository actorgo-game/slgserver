package player

import (
	"github.com/llr104/slgserver/internal/data/model"
	"github.com/llr104/slgserver/internal/protocol"
)

func toRole(r *model.Role) protocol.Role {
	return protocol.Role{
		RId:      r.RId,
		UId:      r.UId,
		NickName: r.NickName,
		Balance:  r.Balance,
		HeadId:   int16(r.HeadId),
		Sex:      r.Sex,
		Profile:  r.Profile,
	}
}

func toRoleRes(r *model.RoleRes) protocol.RoleRes {
	if r == nil {
		return protocol.RoleRes{}
	}
	return protocol.RoleRes{
		Wood:   r.Wood,
		Iron:   r.Iron,
		Stone:  r.Stone,
		Grain:  r.Grain,
		Gold:   r.Gold,
		Decree: r.Decree,
	}
}

func toCity(c *model.MapRoleCity) protocol.MapRoleCity {
	return protocol.MapRoleCity{
		CityId:     c.CityId,
		RId:        c.RId,
		Name:       c.Name,
		X:          c.X,
		Y:          c.Y,
		IsMain:     c.IsMain,
		CurDurable: c.CurDurable,
		MaxDurable: c.MaxDurable,
		OccupyTime: c.OccupyTime.Unix(),
	}
}

func toGeneral(g *model.General) protocol.General {
	skills := make([]*protocol.GSkill, 0, len(g.Skills))
	for _, s := range g.Skills {
		skills = append(skills, &protocol.GSkill{Id: s.Id, Lv: s.Lv, CfgId: s.CfgId})
	}
	return protocol.General{
		Id:            g.Id,
		CfgId:         g.CfgId,
		PhysicalPower: g.PhysicalPower,
		Level:         int8(g.Level),
		Exp:           g.Exp,
		Order:         g.Order,
		CityId:        g.CityId,
		CurArms:       g.CurArms,
		HasPrPoint:    g.HasPrPoint,
		UsePrPoint:    g.UsePrPoint,
		AttackDis:     g.AttackDis,
		ForceAdded:    g.ForceAdded,
		StrategyAdded: g.StrategyAdded,
		DefenseAdded:  g.DefenseAdded,
		SpeedAdded:    g.SpeedAdded,
		DestroyAdded:  g.DestroyAdded,
		StarLv:        g.StarLv,
		Star:          g.Star,
		ParentId:      g.ParentId,
		Skills:        skills,
		State:         g.State,
	}
}

func toArmy(a *model.Army) protocol.Army {
	return protocol.Army{
		Id:       a.Id,
		CityId:   a.CityId,
		Order:    a.Order,
		Generals: a.Generals,
		Soldiers: a.Soldiers,
		ConTimes: a.ConscriptTimes,
		ConCnts:  a.ConscriptCnts,
		Cmd:      a.Cmd,
		FromX:    a.FromX,
		FromY:    a.FromY,
		ToX:      a.ToX,
		ToY:      a.ToY,
		Start:    a.Start,
		End:      a.End,
	}
}

func toBuild(b *model.MapRoleBuild) protocol.MapRoleBuild {
	return protocol.MapRoleBuild{
		RId:        b.RId,
		Name:       b.Name,
		X:          b.X,
		Y:          b.Y,
		Type:       b.Type,
		Level:      b.Level,
		OPLevel:    b.OPLevel,
		CurDurable: b.CurDurable,
		MaxDurable: b.MaxDurable,
		OccupyTime: b.OccupyTime.Unix(),
		EndTime:    b.EndTime.Unix(),
		GiveUpTime: b.GiveUpTime,
	}
}
