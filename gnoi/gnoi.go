package gnoi

//
// func Server(bindAddress string, close <-chan bool, certificates []tls.Certificate, caPool *x509.CertPool) {
// 	opts := []grpc.ServerOption{grpc.Creds(credentials.NewTLS(&tls.Config{
// 		ClientAuth:   tls.RequireAnyClientCert, // tls.NoClientCert, //tls.RequireAndVerifyClientCert, //
// 		Certificates: certificates,
// 		ClientCAs:    caPool,
// 	}))}
//
// 	g := grpc.NewServer(opts...)
// 	pb.RegisterCertificateManagementServer(g, &server{})
// 	reflection.Register(g)
//
// 	log.Infof("starting to listen on %s", bindAddress)
// 	listen, err := net.Listen("tcp", bindAddress)
// 	if err != nil {
// 		log.Exitf("failed to listen: %v", err)
// 	}
//
// 	if close != nil {
// 		go func() {
// 			<-close
// 			g.GracefulStop()
// 		}()
// 	}
//
// 	log.Info("starting to serve")
// 	if err := g.Serve(listen); err != nil {
// 		log.Exitf("failed to serve: %v", err)
// 	}
// }
//
// func Client(targetAddr string, certificates []tls.Certificate, caPool *x509.CertPool) error {
// 	opts := []grpc.DialOption{}
// 	tlsConfig := &tls.Config{}
//
// 	tlsConfig.InsecureSkipVerify = true
// 	tlsConfig.ServerName = "server"
// 	tlsConfig.Certificates = certificates
// 	tlsConfig.RootCAs = caPool
// 	opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
//
// 	conn, err := grpc.Dial(targetAddr, opts...)
// 	if err != nil {
// 		log.Exitf("Dialing to %q failed: %v", targetAddr, err)
// 	}
// 	defer conn.Close()
//
// 	client := pb.NewCertificateManagementClient(conn)
//
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
// 	defer cancel()
//
// 	_, err = client.GetCertificates(ctx, &pb.GetCertificatesRequest{})
// 	if err != nil {
// 		log.Errorf("Hello failed: %v", err)
// 	} else {
// 		log.Info("Received Hello")
// 	}
//
// 	return err
// }
