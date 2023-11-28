Produto x Consumidor

Introdução
O propósito do trabalho é  resolver o problema do gargalo nas requisições da API, garantindo o processamento eficiente de todas as transações.

Metodologia

Para que o problema fosse resolvido foi optado por utilizar a linguagem de programação Go Lang, que nativamente foi projetada para lidar com paralelismo. 
No cliente foi implementado uma função que gera 1000 requisições e armazena em um Json. O envio para a API é feito de forma fragmentada separando o json em Batches acelerando a comunicação entre cliente/servidor.
O processamento das requisições por parte da API é realizado com “Goroutines”, permitindo que várias requisições sejam processadas ao mesmo tempo, para o recebimento dos dados é utilizado um canal “filaRequisicoes” que faz o envio para cada “Goroutines”, dessa forma essa fila organiza as requisições a serem processadas. A implementação da fila também lida com “deeadlock” visto que uma operação de envio só é feita se uma operação de recebimento estiver livre, caso contrário é feito um bloqueio dessa comunicação até que esteja disponível para recebimento.
Outra questão que foi levantada durante o processo de desenvolvimento foi a de “corrida”, onde uma ou mais “goroutines” tentam acessar o mesmo recurso, para resolver isso foi implementado ID exclusivos para cada requisição através de uma requisição atômica.
Para o retorno das informações ao cliente foi optado por fragmentar o arquivo json da mesma forma que foi feito no envio do cliente para a API.







