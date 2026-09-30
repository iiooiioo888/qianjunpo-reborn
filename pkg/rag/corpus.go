package rag

import "context"

// DefaultTopK is the whitepaper retrieve depth for strategic context.
const DefaultTopK = 5

// CorpusEntry is one knowledge chunk with optional faction tag.
type CorpusEntry struct {
	ID       string
	Faction  string
	General  string
	Content  string
	Keywords string
}

// ThreeKingdomsSeed returns a small offline knowledge base (factions / generals / tactics).
func ThreeKingdomsSeed() []CorpusEntry {
	return []CorpusEntry{
		{ID: "wei-cao", Faction: "魏", General: "曹操", Content: "曹操善用奇襲與離間，官渡以少勝多，重視屯田與軍紀。", Keywords: "cao cao wei surprise guandu"},
		{ID: "wei-xiahou", Faction: "魏", General: "夏侯淵", Content: "夏侯淵守合肥，擅長快速援護與騎兵反擊。", Keywords: "xiahou cavalry hefei wei"},
		{ID: "shu-liu", Faction: "蜀", General: "劉備", Content: "劉備以仁德聚人心，入蜀後穩固益州，聯吳抗魏。", Keywords: "liu bei shu benevolence yizhou"},
		{ID: "shu-zhang", Faction: "蜀", General: "張飛", Content: "張飛長坂橋斷後，步兵衝陣兇悍，適合阻敵與斷橋。", Keywords: "zhang fei infantry choke bridge shu"},
		{ID: "shu-guan", Faction: "蜀", General: "關羽", Content: "關羽水軍與陸戰皆強，襄樊之戰展現圍城與斷糧思路。", Keywords: "guan yu naval siege shu"},
		{ID: "wu-sun", Faction: "吳", General: "孫權", Content: "孫權據江東水網，江防與火攻為吳軍核心戰術。", Keywords: "sun quan wu river fire"},
		{ID: "wu-zhou", Faction: "吳", General: "周瑜", Content: "周瑜赤壁火攻，以弱勝強，重視風向與聯合艦隊。", Keywords: "zhou yu chibi fire fleet wu"},
		{ID: "tactic-spear", Faction: "", General: "", Content: "長槍方陣克制騎兵衝擊，需配合側翼弓手。", Keywords: "spear wall anti cavalry archer"},
		{ID: "tactic-flank", Faction: "", General: "", Content: "側翼包抄需隱蔽行軍，待正面接戰後再出擊。", Keywords: "flank cavalry ambush maneuver"},
		{ID: "tactic-hill", Faction: "", General: "", Content: "佔據高地可增弓兵射程並削弱敵衝鋒勢能。", Keywords: "hill archer high ground range"},
	}
}

// SeedStore builds an InMemoryStore from the Three Kingdoms corpus using the given embedder.
func SeedStore(ctx context.Context, embedder Embedder) (*InMemoryStore, error) {
	if embedder == nil {
		embedder = NewHashBagEmbedder(64)
	}
	s := &InMemoryStore{embedder: embedder}
	docs := make([]Document, 0, len(ThreeKingdomsSeed()))
	for _, e := range ThreeKingdomsSeed() {
		text := e.Content + " " + e.Keywords
		if e.Faction != "" {
			text += " faction:" + e.Faction
		}
		if e.General != "" {
			text += " general:" + e.General
		}
		docs = append(docs, Document{
			ID:       e.ID,
			Content:  e.Content,
			Faction:  e.Faction,
			General:  e.General,
			Vector:   embedder.Embed(text),
			RawText:  text,
		})
	}
	if err := s.Upsert(ctx, docs...); err != nil {
		return nil, err
	}
	return s, nil
}
