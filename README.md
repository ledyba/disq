# Nothing-Shared distributed dnsmasq for datacenters

何もデータの共有をしない（〜デバッグが楽）なdnsmasq（のようなもの）。

## disqがしてくれること

 - jsonファイルに書かれた、次の設定について理解する：
   - DNSサーバとして振る舞うべきポートと、応答してもよいネットワーク
   - disqの走るサーバにつかがれた0個以上のipv4ネットワーク
     - netmaskやdefault gatewayなどの、ネットワークに繋がれたコンピュータが知るべき情報
   - disqが面倒を見る0台以上のコンピュータ
     - と、さらにコンピュータに繋がれた0個以上のNICとそれに対応するIPアドレスと対応するFQDN
 - DHCPサーバとして振る舞い、ipアドレスをサーバに割り当てる
 - 複数のdisqが協業し、DHCPとDNSを冗長化して提供する
 - WarningとErrorはZabbixへ通知を行う

# アルゴリズムの概要

## アドレスは事前に与えておく

DHCPサーバですが、コンピュータに割り当てるIPアドレスやホスト名は、事前にjsonファイルに書いて決定します。

それ以外のコンピュータが繋いできた時は、何もしない上でzabbixに通知を投げます（何かしらの不正アクセスである可能性があります）。

## DNSの冗長化

DHCPのメッセージにDNSを複数書けるので、そのままです。特になにもする必要はありません。

正直に言うと、複数書いた時にクライアントがどう冗長化してくれるのかはわからずにこの文章を今書いています。

## DHCPの冗長化

<del>WiFiやイーサネットの衝突回避などと同じアルゴリズムを使います。</del>

<del>すなわち、応答するまでにランダムな待ち時間を入れ、その時間内にリプライを検知したら応答しません。</del>

<del>逆に、その時間内に他のマシンからブロードキャストでリプライが飛んで来なかったら、「自分の番」として応答します。</del>

…と思ったものの、ほぼ同じ内容のメッセージを複数送り返しても問題なさそうなので気にせず送ることにしました。

