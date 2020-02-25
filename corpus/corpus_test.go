package corpus

import (
	"testing"
	"tfidf/similarity"
	"tfidf/util"
)

func TestCorpus(t *testing.T) {
	categories := []string{"新闻", "体育"}
	corpus := NewCorpus("./", categories)
	defer corpus.Free()

	txt, err := util.ReadFileAll("../test/news0.txt")
	if err != nil {
		t.Log("read file failed")
	}
	doc := `
		<p><br/></p><p>连续一个月 每天吃一斤橘子 &nbsp;爱吃橘子的爷爷 变成了“小黄人”</p><p><img src="/img/b3/bd/3f/59d78d2fea2398e949cb611853.jpg"/></p><p>江爷爷手部皮肤比正常人偏黄很多</p><p><img src="/uploads/20181226/1545787532233026.jpeg" title="1545787532233026.jpeg" alt="Unknown-1.jpeg"/></p><p>近来，门口几株橘子树挂起累累硕果，家住天兴洲附近的江佩先生每天都要吃上一斤多橘子，结果全身都变“黄”了。</p><p><img src="/uploads/20181226/1545787544547869.jpeg" title="1545787544547869.jpeg" alt="Unknown.jpeg"/></p><p>22日，82岁江爷爷的外孙小徐一回到家，就发现爷爷的脸色不太对——面色、手部看上去黄黄的，全身皮肤都很晦暗。小徐担心爷爷是肝胆出了问题才有“黄疸”症状，又怕吓着老人家，于是借口带爷爷看湿疹去了医院。</p><p><img src="/uploads/20181226/1545787550935766.jpg" title="1545787550935766.jpg" alt="u=1174518508,3938218017&amp;fm=26&amp;gp=0.jpg"/></p><p>在武汉市中医医院皮肤科，老人全身暗黄的肤色引起了接诊的邱百怡医生的警惕，检查结果显示，江爷爷血液中胆红素指标正常，肝功能也正常,并非“黄疸”。后经详细问诊得知，因家里橘子丰收，老人家竟然持续一个月每天吃上一斤多橘子。</p><p><img src="/uploads/20181226/1545787580233527.jpg" title="1545787580233527.jpg" alt="u=124308378,1542349131&amp;fm=26&amp;gp=0.jpg"/></p><p>邱百怡医生解释，该病被称为“胡萝卜素血症”，因较长时间大量摄入含丰富胡萝卜素的食物之后，血液中的胡萝卜素转运到表皮导致皮肤发黄。只要禁食高胡萝卜素食物一段时间，就可恢复正常。</p><p style="width:100%;text-align:right;color:#999;display: none">责任编辑：薛乔(EK001)</p><p><img src="/uploads/20181226/1545787574223399.jpg" title="1545787574223399.jpg" alt="u=2585472765,264580762&amp;fm=26&amp;gp=0.jpg"/></p>	`
	corpus.AddDocs("新闻", txt)
	wt, _ := corpus.ExtractTags("新闻", doc, 20)
	for _, v := range wt {
		t.Logf("tag: %s | weight: %f", v.Term, v.Weight)
	}
	t.Log("=========================================")
	wt2 := corpus.ExtractWithWeight(doc, 20)
	for _, v := range wt2 {
		t.Logf("word: %s | weight: %f", v.Term, v.Weight)
	}
	t.Log("=========================================")

	m1, _ := corpus.Cal("新闻", doc)
	m2, _ := corpus.Cal("新闻", doc+"测试差异性文本")
	sim := similarity.Cosine(m1, m2)
	t.Logf("sim: %f ", sim)
}
