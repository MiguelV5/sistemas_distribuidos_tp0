# Informe TP0 - Sistemas Distribuidos

---

<br>
<p align="center">
  <img src="https://raw.githubusercontent.com/MiguelV5/MiguelV5/main/misc/logofiubatransparent_partialwhite.png" width="60%"/>
</p>
<br>

---

<br>
<p align="center">
<font size="+3">
Miguel Angel Vasquez Jimenez - 107378
<br>
<br>
Primer cuatrimestre - 2024
</font>
</p>
<br>

---

<br>

A continuación se presentan aclaraciones y notas generales acerca del trabajo practico.


> [!NOTE] 
> Para facilitar la navegación por la solución de los distintos ejercicios, se dividieron los commits en distintas ramas, cada una correspondiente a un ejercicio o subejercicio. Tomar el commit final de cada rama como la solución final del ejercicio correspondiente.
Cada rama surge de la rama de su ejercicio predecesor.

## Ejercicios 1 y 1.1

Se realizó un script de bash para poder parametrizar la cantidad de clientes que se quieren definir en el archivo `docker-compose-dev.yaml` sin tener que modificarlo manualmente.
El mismo reescribe el archivo docker-compose-dev.yaml con la cantidad que se le solicite. 

Ejecución del script:

```bash
./create-multiclient-docker-compose.sh <number_of_clients>
```

## Ejercicio 2

Se modifica el script `create-multiclient-docker-compose.sh` y se genera nuevamente el `docker-compose-dev.yaml` para añadir volumenes (tanto para server como para clientes) de tal forma que los containers puedan utilizar los archivos de configuracion desde el host sin necesidad de estar ligados al build de las imagenes en sí.

Dichos volumenes son declarados con la opción de solo lectura `ro`, e incluyen exclusivamente a los archivos de configuración correspondientes.

Para la ejecución y testeo:
- Se comenta temporalmente el target `docker-image` en el makefile.
- Se ejecuta el target:

```bash
make docker-compose-up
```

- Se cambian las configuraciones y se ejecuta nuevamente para ver reflejados los cambios sin haber hecho build nuevamente.

## Ejercicio 3

Para el test se define una nueva imagen `netcat-tester` que utilice el script `sv-test.sh` en su correspondiente contenedor, que se comunique con el servidor a traves de la misma network definida en `docker-compose-sv-test.yaml`.

Se añadieron adicionalmente los targets `netcat-sv-test-up` y `netcat-sv-test-down` al makefile para correr la prueba facilmente de la forma:

```bash
make netcat-sv-test-up
```

## Ejercicio 4

Se agrega el manejo de señales para `client/common/client.go` y `server/common/server.py` por medio de la recepción de SIGTERM para el cierre adecuado de file descriptors. Particularmente se decidió cerrar los sockets de comunicación al mismo momento de recibir la señal, y adicionalmente para el server, se dejan de aceptar clientes que soliciten conectarse tras haber recibido la misma. 
Tambien se modifica el tiempo que espera docker-compose-down para enviar la señal de terminación, de tal forma que se pueda observar el cierre graceful.

Para la verificación de funcionamiento:

- Iniciar la comunicación:

```bash
make docker-compose-up
```

- Observar los logs en una terminal adicional:

```bash
make docker-compose-logs
```

- Enviar la señal:

```bash
make docker-compose-down
```

- Output:
```bash
...
client2  | time="2024-03-22 21:25:58" level=info msg="action: receive_message | result: success | client_id: 2 | msg: [CLIENT 2] Message N°1\n"
client1  | time="2024-03-22 21:26:03" level=info msg="action: socket_closing | result: success | client_id: 1"
client1  | time="2024-03-22 21:26:03" level=info msg="action: signal_receiver_channel_shutdown | result: success | client_id: 1"
client1  | time="2024-03-22 21:26:03" level=info msg="action: loop_finished | result: success | client_id: 1"
client2  | time="2024-03-22 21:26:03" level=info msg="action: socket_closing | result: success | client_id: 2"
client2  | time="2024-03-22 21:26:03" level=info msg="action: signal_receiver_channel_shutdown | result: success | client_id: 2"
client2  | time="2024-03-22 21:26:03" level=info msg="action: loop_finished | result: success | client_id: 2"
client3  | time="2024-03-22 21:26:03" level=info msg="action: socket_closing | result: success | client_id: 3"
client3  | time="2024-03-22 21:26:03" level=info msg="action: signal_receiver_channel_shutdown | result: success | client_id: 3"
client3  | time="2024-03-22 21:26:03" level=info msg="action: loop_finished | result: success | client_id: 3"
client1 exited with code 0
client2 exited with code 0
client3 exited with code 0
server   | 2024-03-22 21:26:04 INFO     action: exiting_due_to_signal | result: in_progress
server   | 2024-03-22 21:26:04 INFO     action: socket_closing | result: success
server exited with code 0
```

## Ejercicio 5

Se establece como protocolo de comunicación el intercambio de mensajes representados como strings con los siguientes formatos según tipo de mensaje:

- (Cliente -> Servidor) Envio de datos con apuestas. Cada conjunto de parametros encerrado por un par de llaves representa una apuesta con sus respectivos datos separados por comas (`,`). Cada parametro es un par `clave:valor`. La finalización del mensaje completo se denota con el delimitador `;`.

`{PlayerName:STRING,PlayerSurname:STRING,PlayerDocID:INT,PlayerDateOfBirth:STRING,WageredNumber:INT,AgencyID:INT},{ ... }, ... , { ... };`

(En particular para este ejercicio los mensajes solo contienen una apuesta).

- (Servidor -> Cliente) Confirmación de apuesta recibida. Se envia cuando la apuesta es almacenada adecuadamente en el servidor. Este mensaje mantiene el comportamiento de EchoServer, enviando el mismo mensaje del cliente como confirmación. Este tipo de mensaje se define para este ejercicio exclusivamente.

Adicionalmente se definieron las variables de entorno de la apuesta de prueba (`NOMBRE`, `APELLIDO`, `DOCUMENTO`, `NACIMIENTO` y `NUMERO`) en el archivo de configuración.
De esta forma, se puede seguir ejecutando el caso de forma sencilla:

```bash
make docker-compose-up
```

## Ejercicio 6

Se añade toda la logica de comunicación y manejo de mensajes en el servidor y cliente para poder enviar chunks de apuestas de cantidad configurable, por medio del valor de `bets_per_chunk` en `config.yaml`. A su vez se remueve la utilización de las variables `lapse` y `period` en el archivo de configuración, ya que a partir del ejercicio actual imposibilitan la comunicación completa de las apuestas dado el tamaño de los mensajes y el tiempo de espera entre ellos.

Adicionalmente se agrega un mensaje nuevo al protocolo de comunicación (Servidor -> Cliente), que es el mensaje de confirmación de recepción de un chunk de apuestas. Este mensaje se envia una vez que el servidor recibe exitosamente un chunk, para luego almacenarlo en el archivo `bets.csv`.  

Cabe aclarar que se sigue manteniendo el envío de un solo mensaje por conexión, pero se envian chunks de apuestas en cada mensaje. Idealmente se podría mantener la misma conexión para enviar todos los mensajes entre cliente y servidor, pero para el ejercicio actual esto causaría que los clientes solo pudieran conectarse de manera secuencial y tendrian que esperar a que otro cliente termine todos sus envios antes de poder comenzar con los suyos. Esto es solucionable añadiendo manejo concurrente de conexiones en el servidor, sin embargo este es el objetivo del ejercicio 8. 

Para realizar una verificación completa con 5 clientes, ejecutar:

- Iniciar la comunicación:

```bash
make docker-compose-up
```

- Observar los logs en una terminal adicional:

```bash
make docker-compose-logs
```

- Una vez se observe que los clientes han enviado todas las apuestas se puede verificar el tamaño final del archivo bets.csv (78697 lineas) que fue escrito dentro del contenedor del servidor. Esto se puede hacer con el siguiente comando:

```bash
docker exec server wc --lines bets.csv
```

