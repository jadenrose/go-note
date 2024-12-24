DROP TABLE IF EXISTS quick_search;
DROP TABLE IF EXISTS notes;
DROP TABLE IF EXISTS blocks;
DROP TABLE IF EXISTS notes_archive;
DROP TABLE IF EXISTS blocks_archive;

CREATE TABLE IF NOT EXISTS notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    modified_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    title TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS blocks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sort_order INTEGER NOT NULL,
    content TEXT NOT NULL,
    note_id INTEGER NOT NULL,
    FOREIGN KEY (note_id) REFERENCES notes (id)
);

CREATE VIRTUAL TABLE IF NOT EXISTS quick_search
USING fts5(note_id UNINDEXED, title, content, tokenize="trigram");

CREATE TRIGGER IF NOT EXISTS add_note_to_quick_search
AFTER INSERT ON notes
    BEGIN
        INSERT INTO quick_search (note_id, title)
        VALUES (NEW.id, NEW.title);
    END;

CREATE TRIGGER IF NOT EXISTS update_note_in_quick_search
AFTER UPDATE OF title ON notes
    BEGIN
        UPDATE quick_search
        SET title = NEW.title
        WHERE note_id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS remove_note_from_quick_search
AFTER DELETE ON notes
    BEGIN
        DELETE FROM quick_search
        WHERE note_id = OLD.id;
    END;

CREATE TRIGGER IF NOT EXISTS add_block_to_quick_search
AFTER INSERT ON blocks
    BEGIN
        UPDATE quick_search
        SET (content) = (
            SELECT
                group_concat(content, ' | ')
            FROM blocks
            WHERE note_id = NEW.note_id
        )
        WHERE note_id = NEW.note_id;
    END;

CREATE TRIGGER IF NOT EXISTS update_block_in_quick_search
AFTER UPDATE OF content ON blocks
    BEGIN
        UPDATE quick_search
        SET (content) = (
            SELECT
                group_concat(content, ' | ')
            FROM blocks
            WHERE note_id = NEW.note_id
        )
        WHERE note_id = NEW.note_id;
    END;

CREATE TABLE IF NOT EXISTS notes_archive (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME NOT NULL,
    modified_at DATETIME NOT NULL,
    archived_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    title TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS blocks_archive (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sort_order INTEGER NOT NULL,
    content TEXT NOT NULL,
    note_id INTEGER NOT NULL,
    FOREIGN KEY (note_id) REFERENCES notes_archive (id)
);

-- Seed notes and blocks data

-- Insert 20 notes with varying titles
INSERT INTO notes (title) VALUES 
('Note 1: Introduction to the Project'),
('Note 2: Ideas for New Features'),
('Note 3: Client Feedback'),
('Note 4: Meeting Notes - 12/20/2024'),
('Note 5: Research on Competitive Products'),
('Note 6: Todo List for December'),
('Note 7: Project Timeline Discussion'),
('Note 8: Upcoming Deadlines'),
('Note 9: Important Bug Fixes'),
('Note 10: Development Plan - Sprint 1'),
('Note 11: Product Specifications'),
('Note 12: Meeting with Marketing Team'),
('Note 13: Design Ideas for the Homepage'),
('Note 14: Code Review Notes'),
('Note 15: User Stories Breakdown'),
('Note 16: Feature Walkthroughs'),
('Note 17: Database Schema Updates'),
('Note 18: Client Requirements Gathering'),
('Note 19: Final Presentation Slides'),
('Note 20: Deployment Plan for Version 2.0');



-- Insert blocks for each note
-- Note 1
INSERT INTO blocks (sort_order, content, note_id) VALUES 
(1, 'The project aims to streamline workflows for better productivity.', 1),
(2, 'We need to gather initial feedback from the client to refine the approach.', 1);

-- Note 2
INSERT INTO blocks (sort_order, content, note_id) VALUES 
(1, 'Adding user notifications will increase engagement significantly.', 2),
(2, 'We should integrate a task management feature.', 2),
(3, 'The dashboard could have a new look with clearer analytics.', 2);

-- Note 3
INSERT INTO blocks (sort_order, content, note_id) VALUES 
(1, 'Client requested changes to the user interface for easier navigation.', 3),
(2, 'They also suggested adding more customization options for the homepage.', 3);

-- Note 4
INSERT INTO blocks (sort_order, content, note_id) VALUES 
(1, 'Discussed progress on the current sprint and upcoming tasks.', 4),
(2, 'The team agreed to focus on bug fixes first.', 4),
(3, 'Set a deadline for finalizing the product by next week.', 4);

-- Note 5
INSERT INTO blocks (sort_order, content, note_id) VALUES 
(1, 'Research shows that similar products have more advanced reporting features.', 5),
(2, 'They also provide integrations with third-party tools.', 5),
(3, 'We need to consider these factors when designing our product.', 5),
(4, 'Additional features may include real-time collaboration tools.', 5);

-- Note 6
INSERT INTO blocks (sort_order, content, note_id) VALUES 
(1, 'Finish feature updates for the UI by 12/22.', 6),
(2, 'Work on code documentation.', 6);

-- Note 7
INSERT INTO blocks (sort_order, content, note_id) VALUES 
(1, 'Timeline has been adjusted, we need to ensure deadlines are met.', 7),
(2, 'We should allocate more time for user testing in the last phase.', 7),
(3, 'Consider possible delays due to resource allocation.', 7);

-- Note 8
INSERT INTO blocks (sort_order, content, note_id) VALUES 
(1, 'The next major deadline is in 3 days, make sure all tasks are tracked.', 8),
(2, 'Design review is scheduled for 12/22.', 8);

-- Note 9
INSERT INTO blocks (sort_order, content, note_id) VALUES 
(1, 'Fixing the login issue has top priority this week.', 9),
(2, 'Another issue was reported regarding the sign-up flow.', 9),
(3, 'Investigating payment gateway failures.', 9);

-- Note 10
INSERT INTO blocks (sort_order, content, note_id) VALUES 
(1, 'The sprint goals are outlined as follows...', 10),
(2, 'Starting with feature X development.', 10),
(3, 'Allocate 3 days for testing and bug fixing.', 10);

-- Note 11
INSERT INTO blocks (sort_order, content, note_id) VALUES 
(1, 'Define product features for the initial release.', 11),
(2, 'Focus on minimum viable product (MVP) first.', 11),
(3, 'Later updates will introduce additional features and options.', 11);

-- Note 12
INSERT INTO blocks (sort_order, content, note_id) VALUES 
(1, 'Meeting with marketing team went well.', 12),
(2, 'They provided insights into customer demographics.', 12),
(3, 'Discussed marketing strategies and campaign plans.', 12);

-- Note 13
INSERT INTO blocks (sort_order, content, note_id) VALUES 
(1, 'Homepage should have a more modern design.', 13),
(2, 'Consider adding a search feature to improve navigation.', 13),
(3, 'We need to revamp the product description section.', 13);

-- Note 14
INSERT INTO blocks (sort_order, content, note_id) VALUES 
(1, 'The code review for the latest features was positive.', 14),
(2, 'Some minor optimizations were suggested.', 14);

-- Note 15
INSERT INTO blocks (sort_order, content, note_id) VALUES 
(1, 'Break down the user stories into smaller tasks for better management.', 15),
(2, 'Focus on core functionality first, then add enhancements later.', 15);

-- Note 16
INSERT INTO blocks (sort_order, content, note_id) VALUES 
(1, 'Feature walkthrough for the product was successful.', 16),
(2, 'We received feedback on the new features.', 16),
(3, 'Additional improvements will be rolled out next week.', 16);

-- Note 17
INSERT INTO blocks (sort_order, content, note_id) VALUES 
(1, 'Database schema needs to be updated to support new features.', 17),
(2, 'Optimize queries to handle larger datasets.', 17),
(3, 'Consider introducing new relationships for future scalability.', 17);

-- Note 18
INSERT INTO blocks (sort_order, content, note_id) VALUES 
(1, 'Gathering client requirements for the new version of the product.', 18),
(2, 'Need to clarify feature set and prioritize based on client feedback.', 18),
(3, 'Client needs a detailed specification document by 12/25.', 18);

-- Note 19
INSERT INTO blocks (sort_order, content, note_id) VALUES 
(1, 'Final presentation slides are in progress.', 19),
(2, 'Update the product demo with the latest features.', 19),
(3, 'Finalize speaking points for the presentation.', 19);

-- Note 20
INSERT INTO blocks (sort_order, content, note_id) VALUES 
(1, 'Deployment plan involves releasing updates to production next week.', 20),
(2, 'Ensure all pre-release testing is complete before deployment.', 20),
(3, 'Coordinate with the operations team for a smooth rollout.', 20),
(4, 'Monitor post-deployment metrics closely.', 20);

-- Insert 5 archived notes with varying titles
INSERT INTO notes_archive (title, created_at, modified_at) VALUES 
('Note 21: User Interface Enhancements', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('Note 22: API Integration Discussion', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('Note 23: Feature Requests from Users', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('Note 24: Sprint Planning for January', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('Note 25: Post-Launch Analysis', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Note 21
INSERT INTO blocks_archive (sort_order, content, note_id) VALUES 
(1, 'The new UI design will have a more minimalist approach.', 1),
(2, 'Focus on improving the navigation bar layout for easier access.', 1),
(3, 'Ensure the new design is mobile-friendly.', 1),
(4, 'Consider adding user theme customization options.', 1);

-- Note 22
INSERT INTO blocks_archive (sort_order, content, note_id) VALUES 
(1, 'The integration with the third-party API will allow real-time data syncing.', 2),
(2, 'We need to ensure that error handling is robust for API calls.', 2),
(3, 'Discussed the authentication process for external services.', 2),
(4, 'The API response time needs to be optimized for faster performance.', 2);

-- Note 23
INSERT INTO blocks_archive (sort_order, content, note_id) VALUES 
(1, 'Users have requested a dark mode feature for the app.', 3),
(2, 'Adding a feature to allow custom notifications would be beneficial.', 3),
(3, 'Some users have suggested improving the onboarding experience.', 3),
(4, 'Feature to add saved drafts for form submissions is frequently requested.', 3);

-- Note 24
INSERT INTO blocks_archive (sort_order, content, note_id) VALUES 
(1, 'For the January sprint, we should prioritize refactoring old code.', 4),
(2, 'Add new features based on client feedback from the last meeting.', 4),
(3, 'Set a goal to complete at least 5 major user stories in the first 2 weeks.', 4),
(4, 'We need to allocate some time for addressing tech debt.', 4);

-- Note 25
INSERT INTO blocks_archive (sort_order, content, note_id) VALUES 
(1, 'Post-launch analysis indicates that user engagement has significantly improved.', 5),
(2, 'Some bugs have been reported by early users; need to address them quickly.', 5),
(3, 'The feature usage statistics are being collected for future enhancements.', 5),
(4, 'Customer feedback survey results show high satisfaction with new features.', 5);