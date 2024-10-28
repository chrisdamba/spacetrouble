CREATE TABLE IF NOT EXISTS destinations (
                                            id UUID PRIMARY KEY,
                                            name VARCHAR(100) NOT NULL
    );

CREATE TABLE IF NOT EXISTS users (
                                     id UUID PRIMARY KEY,
                                     first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    gender VARCHAR(10) NOT NULL,
    birthday DATE NOT NULL
    );

CREATE TABLE IF NOT EXISTS flights (
                                       id UUID PRIMARY KEY,
                                       launchpad_id VARCHAR(24) NOT NULL,
    destination_id UUID NOT NULL REFERENCES destinations(id),
    launch_date TIMESTAMP NOT NULL
    );

CREATE TABLE IF NOT EXISTS bookings (
                                        id UUID PRIMARY KEY,
                                        user_id UUID NOT NULL REFERENCES users(id),
    flight_id UUID NOT NULL REFERENCES flights(id),
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );

-- Create function to check launches in same week
CREATE OR REPLACE FUNCTION launch_in_same_week(
    p_launchpad_id VARCHAR,
    p_destination_id UUID,
    p_launch_date TIMESTAMP
) RETURNS BOOLEAN AS $$
BEGIN
RETURN NOT EXISTS (
    SELECT 1
    FROM flights f
    WHERE f.launchpad_id = p_launchpad_id
      AND f.destination_id = p_destination_id
      AND DATE_TRUNC('week', f.launch_date) = DATE_TRUNC('week', p_launch_date)
);
END;
$$ LANGUAGE plpgsql;

-- Insert initial destinations
INSERT INTO destinations (id, name) VALUES
                                        ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Mars'),
                                        ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'Moon'),
                                        ('c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', 'Pluto'),
                                        ('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', 'Asteroid Belt'),
                                        ('e0eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', 'Europa'),
                                        ('f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a66', 'Titan'),
                                        ('g0eebc99-9c0b-4ef8-bb6d-6bb9bd380a77', 'Ganymede')
    ON CONFLICT DO NOTHING;