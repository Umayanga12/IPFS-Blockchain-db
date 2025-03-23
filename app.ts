import express, { Request, Response } from 'express';
import cors from 'cors';
import bodyParser from 'body-parser';
import { Gateway, Wallets } from 'fabric-network';
import * as path from 'path';
import * as fs from 'fs';
import dotenv from 'dotenv';

dotenv.config();

const app = express();
const PORT = process.env.PORT || 5000;

app.use(cors());
app.use(bodyParser.json());

// Hyperledger Fabric network configuration
const ccpPath = path.resolve(
    __dirname,
    '..',
    'fabric-samples',
    'test-network',
    'organizations',
    'peerOrganizations',
    'org1.example.com',
    'connection-org1.json'
);

const walletPath = path.join(__dirname, 'wallet');

async function getContract() {
    const ccp = JSON.parse(fs.readFileSync(ccpPath, 'utf8'));
    const wallet = await Wallets.newFileSystemWallet(walletPath);

    const gateway = new Gateway();
    await gateway.connect(ccp, { wallet, identity: 'admin', discovery: { enabled: true, asLocalhost: true } });

    const network = await gateway.getNetwork('mychannel');
    return network.getContract('mycc');
}

// 📌 Register User API
app.post('/register', async (req: Request, res: Response) => {
    try {
        const { id, name, username, email, role, password } = req.body;
        const contract = await getContract();
        await contract.submitTransaction('CreateUser', id, name, username, email, role, password, 'offline');
        res.status(201).json({ message: 'User registered successfully!' });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

// 📌 Login API
app.post('/login', async (req: Request, res: Response) => {
    try {
        const { username, password } = req.body;
        const contract = await getContract();

        const userBytes = await contract.evaluateTransaction('QueryUser', username);
        const user = JSON.parse(userBytes.toString());

        if (!user) {
            return res.status(404).json({ error: 'User not found' });
        }

        if (user.password !== password) {
            return res.status(401).json({ error: 'Invalid credentials' });
        }

        if (user.status === 'online') {
            return res.status(403).json({ error: 'User is already logged in' });
        }

        await contract.submitTransaction('UpdateUserStatus', username, 'online');
        res.status(200).json({ message: 'Login successful', user });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

// 📌 Logout API
app.post('/logout', async (req: Request, res: Response) => {
    try {
        const { username } = req.body;
        const contract = await getContract();

        const userBytes = await contract.evaluateTransaction('QueryUser', username);
        const user = JSON.parse(userBytes.toString());

        if (!user) {
            return res.status(404).json({ error: 'User not found' });
        }

        if (user.status === 'offline') {
            return res.status(400).json({ error: 'User is already logged out' });
        }

        await contract.submitTransaction('UpdateUserStatus', username, 'offline');
        res.status(200).json({ message: 'Logout successful' });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

// 📌 Get User API
app.get('/user/:id', async (req: Request, res: Response) => {
    try {
        const { id } = req.params;
        const contract = await getContract();

        const userBytes = await contract.evaluateTransaction('QueryUser', id);
        const user = JSON.parse(userBytes.toString());

        if (!user) {
            return res.status(404).json({ error: 'User not found' });
        }

        res.status(200).json(user);
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

// 📌 Update User API
app.put('/user/:id', async (req: Request, res: Response) => {
    try {
        const { id } = req.params;
        const { name, username, email, role, password, status } = req.body;
        const contract = await getContract();

        await contract.submitTransaction('UpdateUser', id, name, username, email, role, password, status);
        res.status(200).json({ message: 'User updated successfully!' });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

// 📌 Delete User API
app.delete('/user/:id', async (req: Request, res: Response) => {
    try {
        const { id } = req.params;
        const contract = await getContract();

        await contract.submitTransaction('DeleteUser', id);
        res.status(200).json({ message: 'User deleted successfully!' });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

// Start Express server
app.listen(PORT, () => {
    console.log(`Server running on port ${PORT}`);
});
