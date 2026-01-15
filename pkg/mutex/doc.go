/*
Package mutex provides a distributed locking mechanism to coordinate access
to shared resources across multiple service instances.

It abstracts the underlying storage (e.g., Redis, ETCD) through the Locker interface,
ensuring that business logic remains decoupled from specific infrastructure.
*/
package mutex
