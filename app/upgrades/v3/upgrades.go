package v3

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

const (
	oldEndTime int64 = 1703435178
	newEndTime int64 = 1688220000
)

var addresses = []string{
	"tgrade1s0lankh33kprer2l22nank5rvsuh9ksa4nr6gl",
	"tgrade179skd62nvdvvt440l0krmlj40ewywv4rscgq8z",
	"tgrade1tkgvwuns7l7vkpc0pq2nnjkkdz509vwrzf86sw",
	"tgrade1we8a49nlqr3apex8zxxahh3zf2ye69dy8pcgmv",
	"tgrade1wlagucxdxvsmvj6330864x8q3vxz4x02d0ssjl",
	"tgrade1cemam36qz7le8p0k9gykvkshnvhussphax76mh",
	"tgrade1hcwcxnz5stwnrupf964lzc3txdzgctv5069nzw",
	"tgrade16g4x972lvchrpc7zgtfad3sjqe3nw5njmuk7rp",
	"tgrade1daujfmddygyty3pjsnr9xhz3vxymh6u00krlym",
	"tgrade1lpwnu27qk29sxptphmkw37x0dzreqz34mg25p8",
	"tgrade1grnsmfmhcsl2dllkyyq7qzm9whlnwxzc77ul0t",
	"tgrade1quw3zpwklv3l4ntpfj37c2tx4393ly03tnfc98",
	"tgrade1efp3hmnslju2pn8g2qukp5k5xs028rhppznk67",
	"tgrade1y4v7dcwe5upna6vpgfggrfy23l07r9jdusek5j",
	"tgrade102c8nrsw5wlezdkj9m6rvmx8rrlwf5n0t2yatd",
	"tgrade1cv4leeaavx5lu5n7jgrdklt76rgx2xtd2hlrue",
	"tgrade1qe8uuf5x69c526h4nzxwv4ltftr73v7qt4v8ku",
	"tgrade10wxn2lv29yqnw2uf4jf439kwy5ef00qdms5tvk",
	"tgrade1jdc8qm80m3lvgajuvn36x2nmxfjauclxtyp7rg",
	"tgrade1cxmsyzr90qh85gwgwnvptukhk2tvvhq6t4dr2a",
	"tgrade1k8efqy9seesd0dcvr7207nmmlkfz944p97fypq",
	"tgrade1tvzlc7n05ht0wx8n77a04kkv75yy8dpsfy4d6h",
	"tgrade1utgcen4kj42gs0cpzkqvyvhu2tcp4pvt4gt8m0",
	"tgrade1q3gxkm46daqw48fmnpqu8sdfcedqhnmzleaccr",
	"tgrade1azrgt5aneucrun989pta6jayexnl6lagfcz927",
	"tgrade17h2x3j7u44qkrq0sk8ul0r2qr440rwgjca5y25",
	"tgrade1wa7cr30cpyacj7eznhpvv3rdperwhle0jeec49",
	"tgrade1zkg2tdja965738slnyfxx5kgqprwfl44ecnh3h",
	"tgrade1jplyne08tx0qu77fatnyun8s0u9mtcgwz84zgv",
	"tgrade18xp9dch3k2uxyrz6mdnqd24vmp2na6u55dxwpc",
	"tgrade1js7ezrm55fqgxu3p62d9xn6patjku2z7ne5dvg",
	"tgrade1c9ye54e3pzwm3e0zpdlel6pnavrj9qqvdvmqdq",
	"tgrade1ypwzuhaffvr06ktu0ne6lnm69gxj32qwx2a7lt",
	"tgrade1n3mhyp9fvcmuu8l0q8qvjy07x0rql8q45a9py4",
	"tgrade1kepe077yknqm9kyt63l4zu9rcjla0aku52f7vn",
	"tgrade1c8xa9nxxuvgd32put8qqmd33r29hwuq2ptzh36",
	"tgrade1hnesd8eqjtpu82t89jeqqs74vte440z4y33za6",
	"tgrade1n5s3tepr6a7dr0n4lzjq2x5jqn0a0hqngzn2dv",
	"tgrade1ey2xwu3tfgqxkg3wmrejt6qmn5dx3fl8cserz7",
	"tgrade1admh0ft2553aw6u9hxn7v2vw488r0yyg6u345u",
	"tgrade1u44vteu9tlzhwk7cxfqekgtc7rumlg32vkxgz5",
	"tgrade18nej8s0ykc88hgfumqdvs6kg9c7h0hdqvpalhe",
	"tgrade18uvsa2m93xkewwg60eylvx27c6qfa3675zfsjj",
	"tgrade154cvfyu85tduekt60ga8ydc45lc76w7yy6935n",
	"tgrade1j50e4wwhw332aq922x45p9phc70r7sy44v44y8",
	"tgrade1w8mztnvl55pwmlkgkpaquax6q37n5d2spaadcn",
	"tgrade1fy2s7er0c6uxc8hmnqfgukvkf7xh22s4upgc7u",
	"tgrade1kcdne83mkvygg7guueswnfyfwtsdmewywvnq5q",
	"tgrade1yj87cjq0ent7jnrj9lfffjhht6602dhy0fzlru",
	"tgrade16ycdyzj48pz4nvdprrxkxkq5ax76ksmg5ux6gj",
	"tgrade1a2fa2c4psh39n8mr62w403smnqxxcynxqgfuxs",
	"tgrade10nq2ea7fms8g58fyaqlc2m3thq9kjx5wun6rk9",
	"tgrade17lclxtnwyk64u9nuzfx0d3ljwzddrht0t965ll",
	"tgrade1s3366h2rpwhvlt0w0x49ssyh27778dyztnsz3g",
	"tgrade1q5q2fkxd92n8da8e4ja9mfcl9cesfg7e6l9rud",
	"tgrade1wgkky0dpzufmqxc93lynymfk6uf68005hdh7x2",
	"tgrade1dz487qtggarfaxja70grhs3lgfv02mpn0l9f3j",
	"tgrade1skc8aut895jvg4hdxx7q89sus5x63edeq0mgrk",
	"tgrade1ydw2lp4gcxn8qv09qe8w5qdpgt8qeu30gpf392",
	"tgrade1y4my6z3lgjgw4f7x6wnldpkfagev2wd7hu6vrg",
	"tgrade1zfcmwh56kmz4wqqg2t8pxrm228dx2c6hwwyxfm",
	"tgrade1x20lytyf6zkcrv5edpkfkn8sz578qg5s7azap8",
	"tgrade1vrq3kjq95kkh26vp3g6sfx84xzw654qa4kg2pe",
}

func Addresses() []string {
	return addresses
}

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ak authkeeper.AccountKeeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		for _, addr := range addresses {
			accAddr, err := sdk.AccAddressFromBech32(addr)
			if err != nil {
				return nil, err
			}
			vestingAccount, ok := ak.GetAccount(ctx, accAddr).(*vestingtypes.ContinuousVestingAccount)
			if !ok {
				return nil, fmt.Errorf("cannot cast account %s to vesting account", accAddr)
			}
			if endTime := vestingAccount.GetEndTime(); endTime != oldEndTime {
				return nil, fmt.Errorf("account %s end time is %d instead of %d", accAddr, endTime, oldEndTime)
			}
			vestingAccount.EndTime = newEndTime
			ak.SetAccount(ctx, vestingAccount)
		}
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
