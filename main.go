package main

import (
	"fmt"
	"log"
	"mykubeapp/git"
)

func main() {
	//// Kubeconfig 로드
	//config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	//if err != nil {
	//	panic(err)
	//}
	//
	//// Dump로 내용 출력
	//spew.Dump(config)
	//
	//// 클라이언트 생성
	//clientset, err := kubernetes.NewForConfig(config)
	//if err != nil {
	//	panic(err)
	//}
	//
	//// 네임스페이스 생성
	//ns := "demo-namespace"
	//_, err = clientset.CoreV1().Namespaces().Create(context.TODO(), &corev1.Namespace{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name: ns,
	//	},
	//}, metav1.CreateOptions{})
	//
	//if err != nil {
	//	panic(err)
	//}
	//
	//fmt.Println("✅ 네임스페이스 생성 완료:", ns)

	// GitHub 클라이언트 생성 (토큰 없이 공개 API 사용)
	client := git.NewClient("")

	// 사용자 정보 가져오기
	user, err := client.GetUser("octocat")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("사용자: %s\n", *user.Login)
	fmt.Printf("이름: %s\n", user.GetName())
	fmt.Printf("팔로워: %d\n", user.GetFollowers())

	// 레포지토리 목록 가져오기
	repos, err := client.GetUserRepos("octocat")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n레포지토리 목록:\n")
	for _, repo := range repos {
		fmt.Printf("- %s: %s\n", *repo.Name, repo.GetDescription())
	}
}
