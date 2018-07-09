package gnoi

//
// func wp(f string) string {
// 	return fmt.Sprintf("%s%s", "testdata/", f)
// }
//
// func TestConnection(t *testing.T) {
//
// 	ca, _ := entity.NewEntity(entity.TemplateCA("ca"))
// 	ca.SignWith(ca)
//
// 	ssc, _ := entity.NewEntity(entity.TemplateCA("ssc"))
// 	ssc.SignWith(ssc)
//
// 	serverParent, _ := entity.NewEntity(entity.TemplateCA("server-parent"))
// 	serverParent.SignWith(ca)
// 	clientParent, _ := entity.NewEntity(entity.TemplateCA("client-parent"))
// 	clientParent.SignWith(ca)
//
// 	server, _ := entity.NewEntity(entity.Template("server"))
// 	server.SignWith(serverParent)
// 	client, _ := entity.NewEntity(entity.Template("client"))
// 	client.SignWith(clientParent)
//
// 	clientCerts := []tls.Certificate{*ssc.Certificate}
// 	serverCerts := []tls.Certificate{*ca.Certificate}
//
// 	caPoolClient := x509.NewCertPool()
// 	caPoolServer := x509.NewCertPool()
// 	// caPool.AddCert(ca.Certificate.Leaf)
// 	caPoolClient.AddCert(serverParent.Certificate.Leaf)
// 	caPoolServer.AddCert(clientParent.Certificate.Leaf)
//
// 	close := make(chan bool)
// 	go Server("127.0.0.1:44455", close, serverCerts, nil)
//
// 	time.Sleep(1 * time.Second)
// 	if err := Client("127.0.0.1:44455", clientCerts, nil); err != nil {
// 		t.Error(err)
// 	}
//
// 	close <- true
// }
