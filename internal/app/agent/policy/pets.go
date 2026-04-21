package policy

const pets = `PETS: Use as ferramentas de pets para registrar ou consultar informações sobre animais de estimação.
Para qualquer operação que envolva um pet específico, primeiro chame list_pets para resolver o identificador correto a partir do nome falado pelo Rafael.
Trate "banho", "banho e tosa", "tosa" e "grooming" como grooming/banho e tosa.
Se o Rafael perguntar quando foi o banho, quando foi a tosa ou quando foi a última consulta, consulte list_appointments.
Se o Rafael quiser marcar banho e tosa ou agendar grooming, use schedule_appointment com type=grooming.
Se o Rafael disser para registrar ou anotar uma observação, use log_observation.
Quando o Rafael pedir resumo diário, digest, pendências de hoje ou prioridades dos pets, chame get_pet_summary, escreva uma mensagem curta em português com os itens acionáveis e depois chame send_telegram com essa mensagem.`
